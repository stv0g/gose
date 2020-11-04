import SparkMD5 from 'spark-md5'
import * as asmCrypto from '@tripod/asmcrypto.js'

export async function sha256sum_crypto(file) {

	return file.arrayBuffer().then((ab) => {
		var start = performance.now()

		var digest = crypto.subtle.digest('SHA-256', ab)

		var end = performance.now()
		console.log("sha256 duration", end - start)

		return digest
	})
}

export async function sha256sum(file) {
	var start = performance.now()

	var md = new asmCrypto.Sha256()
	var reader = file.stream().getReader()
	while (true) {
		const {done, value} = await reader.read()
		if (done)
			break

		md = md.process(value)
	}
	var digest = md.finish()
	var digest = asmCrypto.bytes_to_hex(md.result)

	var end = performance.now()
	console.log("sha256 duration", end - start)

	return digest
}

export async function md5sum(file) {
	var start = performance.now()

	var md = new SparkMD5.ArrayBuffer()
	var reader = file.stream().getReader()

	while (true) {
		const {done, value} = await reader.read()
		if (done)
			break

		md = md.append(value)
	}

	var digest = md.end()

	var end = performance.now()
	console.log("md5 duration", end - start)

	return digest
}
