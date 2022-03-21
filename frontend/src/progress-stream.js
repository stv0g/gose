export class ProgressStream extends TransformStream {
    constructor(cb, total) {
        super({
            start() {
                this.started = Date.now();
                this.running = 0;
                this.total = total;
                this.transferred = 0;
                this.speed = 0;
                this.eta = Infinity;

                cb.start(this);
                cb.progress(this);
            }, // required.
            async transform(chunk, controller) {
                controller.enqueue(await chunk);

                this.transferred += chunk.length;
                this.elapsed = Date.now() - this.started;
                this.speed = this.transferred / (this.elapsed / 1e3); // elapsed is expressed in mili-seconds
                this.eta = (this.total - this.transferred) / this.speed;

                cb.progress(this);
            },
            flush() {
                this.elapsed = Date.now() - this.started;

                cb.finish(this);
            }
        });
    }
}
