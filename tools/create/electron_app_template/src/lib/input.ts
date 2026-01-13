

export function isPressingEnter(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key !== 'Enter') {
        return false
    }
    if (e.nativeEvent.isComposing) {
        // isComposing: 中文输入法下，按下回车键时，会触发两次事件，一次是输入法输入，一次是回车键输入

        return false
    }
    return true
}

export function consumeAsPressingEnter(e: React.KeyboardEvent<HTMLInputElement>) {
    if (!isPressingEnter(e)) {
        return false
    }
    e.preventDefault()
    e.stopPropagation()
    return true
}