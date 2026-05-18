fn main() -> Result<(), Box<dyn std::error::Error>> {
    let protoc = protoc_bin_vendored::protoc_bin_path()?;
    std::env::set_var("PROTOC", protoc);

    tonic_prost_build::configure()
        .build_server(false)
        .compile_protos(&["../../api/proto/webhook.proto"], &["../../api/proto"])?;
    println!("cargo:rerun-if-changed=../../api/proto/webhook.proto");
    Ok(())
}
