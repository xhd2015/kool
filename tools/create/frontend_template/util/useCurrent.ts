import { useRef } from "react"

export function useCurrent<T>(value: T) {
    const ref = useRef(value)
    ref.current = value
    return ref
}
