fn main() {
    println!("cargo:rerun-if-changed=proto/inventory/inventory_service.proto");
    println!("cargo:rerun-if-changed=proto/inventory/inventory_messages.proto");
}
