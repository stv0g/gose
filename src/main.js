import { ProgressBar } from './progress-bar.js'
import { S3Upload } from './s3-upload.js'

import * as utils from './utils.js'
import * as csum from './checksum.js'

export function load() {
	if (!utils.supportsRequestStreams)
		console.log('Seems like we dont support request streams')

	// get a reference to the inputElement in any way you choose
	const inputElm = document.getElementById('file')
	const progressElm = document.getElementById('progress')
	const progressStatsElm = document.getElementById('progress-stats')
	const uploadBtn = document.getElementById('upload')
	const resultElm = document.getElementById("result")

	const statsTransferred = document.getElementById('stats-transferred')
	const statsElapsed = document.getElementById('stats-elapsed')
	const statsEta = document.getElementById('stats-eta')
	const statsSpeed = document.getElementById('stats-speed')

	const progress = new ProgressBar(progressElm)

	inputElm.addEventListener('change', (e) => {
		const file = inputElm.files[0]

		if (!file)
			return

		console.log('File selected', file)
		csum.sha256sum(file)
			.then(digest => {
				console.log('sha256 = ', digest)
			})
	})

	uploadBtn.addEventListener('click', (e) => {
		const file = document.getElementById('file').files[0]
		if (!file)
			return

		let upl = new S3Upload({
			start(p) {
				console.log("Transfer started")
				progress.setMinMax(0, p.total)
				progress.set(p.transferred)
			},
			finish(p) {
				console.log("Transform finished")
				progress.set(0) // reset pbar
			},
			progress(p) {
				progress.set(p.transferred)
				statsTransferred.innerHTML = p.transferred
				statsElapsed.innerHTML = p.elapsed
				statsEta.innerHTML = p.eta
				statsSpeed.innerHTML = p.speed
			}
		})

		upl.upload(file)
			.then((r) => {
				console.log("Upload succeeded")
				console.log(r)

				result.classList.add('alert-success')
				result.innerHTML = `Upload complete: <a href="${r.url}">${r.url}</a>`
			})
			// (r) => {
			// 	console.log("Upload failed")
			// 	console.log(r)

			// 	result.classList.add('alert-danger')
			// 	result.innerHTML = `${r.status} - ${r.statusText}`
			// })

		e.preventDefault()
	})
}
