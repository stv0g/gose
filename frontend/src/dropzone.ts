
export class Dropzone {
    element: HTMLDivElement;
    canDrop: (ev: DragEvent) => boolean;
    handleDrop: (ev: DragEvent) => void;

    constructor(elm: HTMLDivElement, canDrop: (ev: DragEvent) => boolean, handleDrop: (ev: DragEvent) => void) {
        this.element = elm;
        this.canDrop = canDrop;
        this.handleDrop = handleDrop;

        window.addEventListener("dragenter", (ev: DragEvent) => this.showDropZone(ev));
        this.element.addEventListener("dragenter", (ev: DragEvent) => this.allowDrag(ev));
        this.element.addEventListener("dragover", (ev: DragEvent) => this.allowDrag(ev));
        this.element.addEventListener("drop", (ev: DragEvent) => this.handleDrop(ev));
        this.element.addEventListener("dragleave", () => this.hideDropZone());
    }

    protected showDropZone(ev: DragEvent) {
        if (!this.canDrop(ev)) {
            return;
        }
        
        this.element.style.display = "block";
    }

    protected hideDropZone() {
        this.element.style.display = "none";
    }

    protected allowDrag(ev: DragEvent) {
        if (!this.canDrop(ev)) {
            return;
        }

        ev.preventDefault();
        ev.dataTransfer.dropEffect = "copy";
    }
}
