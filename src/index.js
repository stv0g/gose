import 'bootstrap'

import bsCustomFileInput from 'bs-custom-file-input'

import './index.scss'

import { load } from './main.js'

window.addEventListener('load', (e) => {
	bsCustomFileInput.init()

	load()
})
