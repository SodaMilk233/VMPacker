use std::process;

#[inline(never)]
#[unsafe(no_mangle)]
pub extern "C" fn check_key(input: u64) -> u64 {
    ((input * 7) + 42) ^ 0xFF
}

fn main() {
    let result = check_key(10);
    print!("checkKey(10) = {}\n", result);
    if result == 143 {
        print!("[+] OK\n");
        process::exit(0);
    } else {
        print!("[-] FAIL\n");
        process::exit(1);
    }
}
