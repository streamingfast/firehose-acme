use std::{
    env,
    fs::{self, File},
    io::{self, BufWriter},
    os::{self, unix::prelude::FromRawFd},
};

use std::io::Write;

fn main() {
    let payload_file = env::var("PAYLOAD").expect("PAYLOAD environment variable must be set");
    let content = fs::read_to_string(payload_file).expect("unable to read PAYLOAD file");

    let lines: Vec<_> = content.lines().collect();

    let stdout = unsafe { File::from_raw_fd(1) };
    let mut writer = BufWriter::with_capacity(64 * 1024, stdout);

    let new_line = vec!['\n' as u8];

    loop {
        for line in &lines {
            writer.write_all(line.as_bytes());
            writer.write_all(new_line.as_ref());
        }
    }
}
