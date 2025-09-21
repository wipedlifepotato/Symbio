#![no_main]
static mut XOR_KEY: u8 = 0;

#[no_mangle]
pub extern "C" fn set_key(ptr: *mut u8, len: usize) {
    if len == 0 { return; }
    let data = unsafe { std::slice::from_raw_parts(ptr, len) };
    let key = data[len - 1]; // последний байт массива как ключ
    unsafe { XOR_KEY = key; }

    #[cfg(debug_assertions)]
    println!("set_key: XOR_KEY set to {}", key);
}

#[no_mangle]
pub extern "C" fn encrypt_with_key(ptr: *mut u8, len: usize) {
    if len == 0 { return; }
    let data = unsafe { std::slice::from_raw_parts_mut(ptr, len) };
    let key = unsafe { XOR_KEY };

    for b in data.iter_mut() {
        *b ^= key;
    }

    #[cfg(debug_assertions)]
    println!("encrypt_with_key: data after xor = {:?}", data);
}

#[no_mangle]
pub extern "C" fn plugin_entry(ptr: *mut u8, len: usize) {
    let data = unsafe { std::slice::from_raw_parts_mut(ptr, len) };
    if len == 0 {
        return;
    }
    let key = data[len - 1];
  //  unsafe { XOR_KEY = key; }
    for b in &mut data[..len - 1] {
        *b ^= key;
    }
}

