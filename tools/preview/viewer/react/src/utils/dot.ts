import { Graphviz } from "@hpcc-js/wasm/graphviz";

let graphvizInstance: Graphviz | null = null;

async function getGraphviz(): Promise<Graphviz> {
    if (!graphvizInstance) {
        graphvizInstance = await Graphviz.load();
    }
    return graphvizInstance;
}

export async function renderDotToSvg(code: string): Promise<string> {
    const graphviz = await getGraphviz();
    return graphviz.dot(code);
}
