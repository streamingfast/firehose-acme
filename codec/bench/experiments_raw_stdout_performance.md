## Stdout Raw Performance

While exploring Aptos performance, I ran small experiment to gather what is the theoretical throughput we can get from launching a process that writes to `stdout` and having the "supervisor" process read from the stream of data.

### Test Setup

I created two "feeder" program, one in Go:

```go
package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/streamingfast/cli"
)

func main() {
	data := os.Getenv("PAYLOAD")
	cli.Ensure(data != "", "PAYLOAD environment variable must be set")

	content, err := os.ReadFile(data)
	cli.NoError(err, "unable to read payload file %q", data)

	lines := bytes.Split(content, []byte("\n"))

	for {
		for _, line := range lines {
			fmt.Println(line)
		}
	}
}
```

And the other one in Rust:

```rust
use std::{env, fs};

fn main() {
    let payload_file = env::var("PAYLOAD").expect("PAYLOAD environment variable must be set");
    let content = fs::read_to_string(payload_file).expect("unable to read PAYLOAD file");

    let lines: Vec<_> = content.lines().collect();

    loop {
        for line in &lines {
            println!("{}", line);
        }
    }
}
```

To measure the actual throughput of each method, we are going to use [pv (a.k.a Pipe Viewer)](http://www.ivarch.com/programs/pv.shtml) which will consume the output and give average renders of how many bytes pass through the pipe.

### Golang

Invocation is:

```shell
go build -o /tmp/feeder-go ./codec/bench/feeder-go && PAYLOAD=./codec/testdata/full.firelog /tmp/feeder-go | pv -a > /dev/nul
```

And with the original snippet, on my machine, I was able to achieve around `185MiB/s`. Now, this seem low as my assumption was that much more throughput could be achievable.

It appears the Golang `stdout` is not buffered leading to excessive system calls behind made to flush the buffer on each invocation of `Println` and it's what is causing the slow throughput. Changing the Go snippet above to become:

```
	bufferedStdout := bufio.NewWriter(os.Stdout)
	defer bufferedStdout.Flush()

	for {
		for _, line := range lines {
			fmt.Fprintln(bufferedStdout, line)
		}
	}
```

Gives use now `244MiB/s` of throughput. This however is still under achieving it. Let's get rid of `fmt.Fprintln` call and write directly to the writer instead:

```
	bufferedStdout := bufio.NewWriter(os.Stdout)
	defer bufferedStdout.Flush()

	for {
		for _, line := range lines {
			bufferedStdout.Write(append(line, '\n'))
		}
	}
```

Gives use now `1.59GiB/s`, now we start to talk about good throughput here. We can improve the example by tweaking the buffer size a bit, let's use 64KB of buffer:


```
	bufferedStdout := bufio.NewWriterSize(os.Stdout, 64*1024)
	defer bufferedStdout.Flush()

	for {
		for _, line := range lines {
			bufferedStdout.Write(append(line, '\n'))
		}
	}
```

Now we get a whooping `2.19GiB/s` throughput. Can we do it even better? Last thing that hinders performance in this sample is the `append` line. I thought there would be a least one byte more in the `line` buffer so it would allow just to append a byte to the buffer.


```
	bufferedStdout := bufio.NewWriterSize(os.Stdout, 64*1024)
	defer bufferedStdout.Flush()

	for {
		for _, line := range lines {
			bufferedStdout.Write(line)
			bufferedStdout.WriteByte('\n')
		}
	}
```

This gives us what I think is the best we can achieve at `5.75GiB/s` which is roughly 31x better than our initial case.

Of course, this is a synthetic case but it means that if the console reader is not a bottleneck, changing our Geth instrumentation to use buffered output and tweaking how we write to it could achieve a good performance boosts, especially in reprocessing mode where much data is ingested.

> Caveats 1: Flushing will be important, flush will need to happen at block end to be sure console reader can fully read a block and not wait until next block starts

> Caveats 2: Is it a problem if the Geth crashes in the middle and the buffer was not flushed? Is it possible to create a hole here?

> Caveats 3: What about Interrupt signal that can happen as we saw on system (happened on EOSIO in C++), are we correctly covered from that case? Is Go handling of that correct? Hard to tell but warrant some further investigation.

### Rust

Invocation is:

```shell
rm -rf /tmp/feeder-rs && cargo build --release --manifest-path=./codec/bench/feeder-rs/Cargo.toml && cp ./codec/bench/feeder-rs/target/release/feeder-rs /tmp/feeder-rs && PAYLOAD=./codec/testdata/full.firelog /tmp/feeder-rs | pv -a > /dev/null
```

And with the original snippet, on my machine, I was able to achieve around `245MiB/s`. Now, we need seeing Golang that this is quite low, so let's try to bring up the same performance optimization here.

Let's go straight to what worked best for Golang, buffered output, 64KB cache, write directly to buffer.


```
    let stdout = unsafe { File::from_raw_fd(1) };
    let mut writer = BufWriter::with_capacity(64 * 1024, stdout);

    let new_line = vec!['\n' as u8];

    loop {
        for line in &lines {
            writer.write_all(line.as_bytes());
            writer.write_all(new_line.as_ref());
        }
    }
```

With that code, we were able to achieve a whooping 8.86GiB/s which is impressive. Adding error handling to both `write_all` gives use a good `7.82GiB/s` which is pretty solid.

### Read Setup

Now that we have theoretical numbers about how much we can actually write, let's explore how fast we can read that back now from our Golang program.

All test tests below will be running using the `feeder-rs` program so ensure it's compiled:

```
rm -rf /tmp/feeder-rs && cargo build --release --manifest-path=./codec/bench/feeder-rs/Cargo.toml && cp ./codec/bench/feeder-rs/target/release/feeder-rs /tmp/feeder-rs
```

#### Raw Standard Input

Let's see how fast can we just read from the stream. Reading code is more or less like that:

```
	buffer := make([]byte, 64*1024)

	stdOutBytesCounter := dmetrics.NewLocalRateCounter(time.Second, "bytes")

	for {
		count, err := os.Stdin.Read(buffer)
		cli.NoError(err, "Unable to read stdin")

		stdOutBytesCounter.IncBy(int64(count))
	}
```

We run it like this:

```
PAYLOAD=./codec/testdata/full.firelog /tmp/feeder-rs | go run ./codec/bench "-" viaStdinBinary
```

With this, I was able achieve `6.36GiB/s` in average for a 60s run.

#### Raw `bufio.Scanner` (lines based)

Let's switch to a line based approach which is more what we do:

```
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 64*1024), 100*1024*1024)

	stdOutBytesCounter := dmetrics.NewLocalRateCounter(time.Second, "bytes")
	for scanner.Scan() {
		stdOutBytesCounter.IncBy(int64(len(scanner.Bytes())))
	}
```

We run it like this:

```
PAYLOAD=./codec/testdata/full.firelog /tmp/feeder-rs | go run ./codec/bench "-" viaStdinLines
```

In this setup, we were able to achieve `2.51GiB/s` of read throughput in average over 60s run.

### From process Golang with `scanner`

Now instead of consuming through `stdin`, let's actually launch the process and try to consumes its `stdout` output and see what we can get.

There is a `io.Pipe` that is used, the `writer` the part of the pipe is connected to the launched process' `stdout` channel and the other half connected to a `bufio.Scanner` initialized just like in the [raw `bufio.Scanner` lines based](#raw-bufioscanner-lines-based) example.

We run it like:

```
PAYLOAD=./codec/testdata/full.firelog go run ./codec/bench /tmp/feeder-rs viaGolang
```

And a 60s run gave us an average of `2.18GiB/s`. It seems the pipe I/O system is not reading as fast as it could.

### From process Golang with `Overseer`

The `overseer` library is used to manage the process, it's what used in the full system. It has its own tokenizer to determine the max line length.

We run it like:

```
PAYLOAD=./codec/testdata/full.firelog go run ./codec/bench /tmp/feeder-rs viaOverseer
```

And a 60s run gave us an average of `593.31 MiB/s`. A big drop in performance.
