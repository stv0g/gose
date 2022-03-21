import SHA256 from 'crypto-js/sha256';
import MD5 from 'crypto-js/md5';
import WordArray from 'crypto-js/lib-typedarrays';
import { buf2base64, buf2hex } from "./utils";

function WordArrayToUint8Array(wordArray) {
    const l = wordArray.sigBytes;
    const words = wordArray.words;
    const result = new Uint8Array(l);
    var i = 0 /*dst*/ ,
        j = 0 /*src*/ ;
    while (true) {
        // here i is a multiple of 4
        if (i == l)
            break;
        var w = words[j++];
        result[i++] = (w & 0xff000000) >>> 24;
        if (i == l)
            break;
        result[i++] = (w & 0x00ff0000) >>> 16;
        if (i == l)
            break;
        result[i++] = (w & 0x0000ff00) >>> 8;
        if (i == l)
            break;
        result[i++] = (w & 0x000000ff);
    }
    return result;
}

export async function sha256sum(blob) {
    if (window.crypto.subtle !== undefined)
        return await crypto.subtle.digest('SHA-256', blob);
    else {
        let wa = WordArray.create(blob);
        return WordArrayToUint8Array(SHA256(wa))
    }
}

export async function md5sum(blob) {
    const wa = WordArray.create(blob);
    const hash = MD5(wa);
    return WordArrayToUint8Array(hash)
}