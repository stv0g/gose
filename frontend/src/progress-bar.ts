export class ProgressBar {
    elm: HTMLElement
    min: number
    max: number
    value: number

    constructor(elm: HTMLElement) {
        this.elm = elm;
    }

    setMinMax(min: number, max: number) {
        this.min = min;
        this.max = max;

        this.elm.setAttribute("aria-valuemin", this.min.toString());
        this.elm.setAttribute("aria-valuemax", this.max.toString());
    }

    set(val: number) {
        let percent = Number(100 * this.value / this.max);

        this.value = val;

        if (percent > 0) {
            this.elm.textContent = `${percent.toFixed(0)} %`;
        }
    
        this.elm.setAttribute("aria-valuenow", this.value.toString());
        this.elm.setAttribute("style", `width: ${percent}%`);
    }
}
