import * as SHA256 from "crypto-js/sha256";
import * as MD5 from "crypto-js/md5";
import * as WordArray from "crypto-js/lib-typedarrays";

function wordToUintArray(wordArray: WordArray) {
    const l = wordArray.sigBytes;
    const words = wordArray.words;
    const result = new Uint8Array(l);
    var i = 0 /*dst*/ ,
        j = 0 /*src*/ ;
    for (;;) {
        // here i is a multiple of 4
        if (i === l) {
            break;
        }
        var w = words[j++];
        result[i++] = (w & 0xff000000) >>> 24;
        if (i === l) {
            break;
        }
        result[i++] = (w & 0x00ff0000) >>> 16;
        if (i === l) {
            break;
        }
        result[i++] = (w & 0x0000ff00) >>> 8;
        if (i === l) {
            break;
        }
        result[i++] = (w & 0x000000ff);
    }
    return result;
}

export async function sha256sum(buf: ArrayBuffer): Promise<Uint8Array> {
    if (window.crypto.subtle !== undefined) {
        let ab = await crypto.subtle.digest("SHA-256", buf);
    }

    let wa = WordArray.create(buf as unknown as number[]);
    return wordToUintArray(SHA256(wa));
}

export async function md5sum(buf: ArrayBuffer): Promise<Uint8Array> {
    const wa = WordArray.create(buf as unknown as number[]);
    const hash = MD5(wa);
    return wordToUintArray(hash);
}
