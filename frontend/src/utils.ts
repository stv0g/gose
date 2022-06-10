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

export function shortUrl(url: string, l?: number){
    var l = typeof(l) != "undefined" ? l : 32;
    var chunk_l = (l/2);
    var url = url.replace("http://","").replace("https://","");

    if (url.length <= l) {
        return url;
    }

    var start_chunk = shortString(url, chunk_l, false);
    var end_chunk = shortString(url, chunk_l, true);
    return start_chunk + ".." + end_chunk;
}

function shortString(s: string, l: number, reverse?: boolean){
    var stop_chars = [' ','/', '&'];
    var acceptable_shortness = l * 0.80; // When to start looking for stop characters
    var reverse = typeof(reverse) != "undefined" ? reverse : false;
    var s = reverse ? s.split("").reverse().join("") : s;
    var short_s = "";

    for (var i=0; i < l-1; i++) {
        short_s += s[i];
        if (i >= acceptable_shortness && stop_chars.indexOf(s[i]) >= 0) {
            break;
        }
    }

    if (reverse) {
        return short_s.split("").reverse().join("");
    }

    return short_s;
}
