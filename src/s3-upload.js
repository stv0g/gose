import { noConflict } from 'jquery'
import { ProgressStream } from './progress-stream.js'

import * as utils from './utils.js'

const apiBase = '/api/v1'

export class S3Upload {

	constructor(cb) {
		this.mpuParts = []

		this.mpuEnabled = false
		this.mpuThreshold = 1000*1024*1024	// 10 MiB
		this.mpuPartSize  = 5*1024*1024		// 5 MiB
		this.mpuMaxPartSize = 100*1024*1024	// 100 MiB

		this.callbacks = cb
	}

	wrapFile(file) {
		if (utils.supportsRequestStreams) {
			let ps = new ProgressStream(this.callbacks, file.size)

			return file.stream().pipeThrough(ps)
		}
		else
			return file
	}

	async upload(file) {
		let key = utils.uuidv4() + '/' + file.name

		let wrappedFile = this.wrapFile(file)

		return this.mpuEnabled && file.size > this.mpuThreshold
			? this.uploadParts(key, wrappedFile)
			: this.uploadPart(key, wrappedFile)
	}

	async getPresignedUrl(url, method='PUT', uploadId=null) {
		return fetch(apiBase + '/presign/' + url + '?method=' + method, {
			'method': 'GET'
		})
			.then(r => r.json())
			.then(j => j['url'])
	}

	async uploadParts(key, file) {
		let uploadId = this.mpuInitiate(key)

		let parts = []

		let no
		let pos
		for (no = 0, pos = 0; pos + this.mpuPartSize < file.size; no++, pos += this.mpuPartSize) {
			let blob = file.size(pos, pos + this.mpuPartSize)

			let part = await this.uploadPart(key, blob, {
				partNo: no,
				uploadId: uploadId
			})

			parts.push(part)
		}

		{ // Final part
			let part = await this.uploadPart(key, blob, {
				uploadId: uploadId,
				partNo: no
			})

			parts.push(part)
		}

		this.mpuComplete(key, uploadId, parts)
	}

	async uploadPart(key, part) {
		let url = await this.getPresignedUrl(key, 'PUT')

		return await fetch(url, {
			'method': 'PUT',
			'body': part
		})
		.then((r) => {
			return {
				url: url,
				response: r
			}
		})
		.catch((r) => {
			throw r
		})
	}

	mpuInitiate() {
		// TODO
	}

	mpuComplete() {
		// TODO
	}

	mpuAbort() {
		// TODO
	}
}
