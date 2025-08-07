declare module 'plantuml-encoder' {
    export function encode(plantumlCode: string): string;
    export function decode(encodedString: string): string;
}