export function buf2hex(buf: ArrayBuffer): string {
    return Array.prototype.map.call(new Uint8Array(buf), (x: number) => ("00" + x.toString(16)).slice(-2)).join("");
}

export function hex2buf(hex: string): ArrayBuffer {
    let a = new Uint8Array(hex.match(/.{1,2}/g).map(byte => parseInt(byte, 16)));
    return a.buffer;
}


export function arraybufferEqual(a: ArrayBuffer, b: ArrayBuffer) {
    if (a === b) {
      return true;
    }
  
    if (a.byteLength !== b.byteLength) {
      return false;
    }
  
    var view1 = new DataView(a);
    var view2 = new DataView(b);
  
    var i = a.byteLength;
    while (i--) {
        if (view1.getUint8(i) !== view2.getUint8(i)) {
            return false;
        }
    }
  
    return true;
}
