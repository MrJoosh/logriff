use clap::{error::ErrorKind, Parser};
use logriff::cli::{validate_args, Args};
use std::process::ExitCode;

fn main() -> ExitCode {
    // TODO: parse flags, stream input, apply filters, aggregate metrics, and emit output.
    // See README.md for the full CLI specification and output formats.
    let args = match Args::try_parse() {
        Ok(args) => args,
        Err(err) => match err.kind() {
            ErrorKind::DisplayHelp | ErrorKind::DisplayVersion => {
                let _ = err.print();
                return ExitCode::SUCCESS;
            }
            _ => {
                eprintln!("{:?}: {}", err.kind(), err);
                return ExitCode::FAILURE;
            }
        },
    };

    if let Err(msg) = validate_args(&args) {
        eprintln!("{msg}");
        return ExitCode::FAILURE;
    };

    println!("Args: {:?}", args);

    ExitCode::SUCCESS
}
