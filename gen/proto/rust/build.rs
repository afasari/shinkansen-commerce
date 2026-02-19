use std::env;
use std::path::PathBuf;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let out_dir = PathBuf::from(env::var("OUT_DIR")?);

    tonic_build::configure()
        .build_server(true)
        .build_client(true)
        .file_descriptor_set_path(out_dir.join("file_descriptor_set.bin"))
        .out_dir(&std::path::PathBuf::from("src"))
        .compile(
            &[
                "../../../proto/inventory/inventory_service.proto",
                "../../../proto/inventory/inventory_messages.proto",
            ],
            &["../../../proto", "../../../proto/google/api"],
        )?;

    println!("cargo:rerun-if-changed=../../../proto/inventory/inventory_service.proto");
    println!("cargo:rerun-if-changed=../../../proto/inventory/inventory_messages.proto");

    Ok(())
}
