// Check if a file is editable based on file extension
export function isEditableFile(filePath: string): boolean {
    const editableExtensions = ['.md', '.uml', '.puml', '.mmd', '.txt', '.json', '.yaml', '.yml'];
    const ext = filePath.toLowerCase().substring(filePath.lastIndexOf('.'));
    return editableExtensions.includes(ext);
} 