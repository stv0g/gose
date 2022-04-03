type Line = {
    length: number,
    angle: number
};

type Point = number[];

type Bounds = {
    xMin: number,
    xMax: number,
    yMin: number,
    yMax: number
};

export class Chart {
    element: HTMLDivElement;
    smoothing: number = 0.15;
    options: Bounds;

    constructor(elm: HTMLDivElement) {
        this.element = elm;
    }

    protected pointsPositions(points: Point[], bounds: Bounds): Point[] {
        return points.map(e => {
            const map = (value: number, inMin: number, inMax: number, outMin: number, outMax: number): number => {
                return (value - inMin) * (outMax - outMin) / (inMax - inMin) + outMin;
            };

            const x = map(e[0], bounds.xMin, bounds.xMax, -1, 101)
            const y = map(e[1], bounds.yMin, bounds.yMax, 27, 3)

            return [x, y];
        })
    }

    protected line(pointA: Point, pointB: Point): Line {
        const lengthX: number = pointB[0] - pointA[0];
        const lengthY: number = pointB[1] - pointA[1];

        return {
            length: Math.sqrt(Math.pow(lengthX, 2) + Math.pow(lengthY, 2)),
            angle: Math.atan2(lengthY, lengthX)
        }
    }

    protected controlPoint(current: Point, previous: Point, next: Point, reverse: boolean = false): Point {
        const p = previous || current;
        const n = next || current;
        const l = this.line(p, n);

        const angle = l.angle + (reverse ? Math.PI : 0);
        const length = l.length * this.smoothing;
        const x = current[0] + Math.cos(angle) * length;
        const y = current[1] + Math.sin(angle) * length;

        return [x, y];
    }

    protected bezierCommand(point: Point, i: number, a: Point[]): string {
        const cps = this.controlPoint(a[i - 1], a[i - 2], point);
        const cpe = this.controlPoint(point, a[i - 1], a[i + 1], true);
        const close = i === a.length - 1 ? ' z':'';

        return `C ${cps[0]},${cps[1]} ${cpe[0]},${cpe[1]} ${point[0]},${point[1]}${close}`;
    }

    protected svg(points: Point[]) {
        const d = points.reduce((acc, e, i, a) => i === 0
            ? `M ${a[a.length - 1][0]},100 L ${e[0]},100 L ${e[0]},${e[1]}`
            : `${acc} ${this.bezierCommand(e, i, a)}`
        , '');
    
        return `<svg version="1.1" xmlns="http://www.w3.org/2000/svg" viewBox="0 1 100 30"><path d="${d}" /></svg>`;
    }

    protected findBounds(points: Point[]): Bounds {
        let options = {
            xMin: 0,
            xMax: 0,
            yMin: 0,
            yMax: 0
        };
        
        for (let p of points) {
            if (p[1] > options.yMax)
                options.yMax = p[1];
            if (p[1] < options.yMin)
                options.yMin = p[1];
    
            if (p[0] > options.xMax)
                options.xMax = p[0];
            if (p[0] < options.xMin)
                options.xMin = p[0];
        }

        return options;
    }

    public render(points: Point[]) {
        if (points.length <= 1) {
            return
        }

        const bounds = this.findBounds(points);
        const pointsPositions = this.pointsPositions(points, bounds);

        this.element.innerHTML = this.svg(pointsPositions);
    }
}
