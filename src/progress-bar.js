export class ProgressBar {
	constructor(elm, val, min, max) {
		this.elm = elm

		this.setMinMax(min, max)
		this.set(val)
	}

	setMinMax(min, max) {
		this.min = min
		this.max = max

		this.elm.setAttribute('aria-valuemin', this.min)
		this.elm.setAttribute('aria-valuemax', this.max)
	}

	set(val, min, max) {
		this.val = val

		this.elm.setAttribute('aria-valuenow', this.val)
		this.elm.setAttribute('style', 'width:' + Number(100 * this.val / this.max) + '%')
	}
}

