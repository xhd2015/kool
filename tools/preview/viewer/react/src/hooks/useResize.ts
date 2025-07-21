import { useState, useEffect, useCallback } from 'react';
import type { RefObject } from 'react';

interface UseResizeProps {
    containerRef: RefObject<HTMLDivElement | null>;
    direction: 'horizontal' | 'vertical';
    minSize?: number;
    maxSize?: number;
    defaultSize?: number;
    enabled?: boolean;
}

export const useResize = ({
    containerRef,
    direction,
    minSize = 25,
    maxSize = 75,
    defaultSize = 50,
    enabled = true
}: UseResizeProps) => {
    const [size, setSize] = useState<number>(defaultSize); // percentage (0-100)
    const [isDragging, setIsDragging] = useState(false);

    // Handle mouse down - returns the event handler for onMouseDown
    const handleMouseDown = useCallback((e: React.MouseEvent) => {
        if (!enabled) return;
        e.preventDefault();
        setIsDragging(true);
    }, [enabled]);

    // Handle mouse move for resizing
    const handleMouseMove = useCallback((e: MouseEvent) => {
        if (!enabled || !isDragging || !containerRef.current) return;

        const containerRect = containerRef.current.getBoundingClientRect();
        const containerSize = direction === 'horizontal' ? containerRect.width : containerRect.height;
        const currentPos = direction === 'horizontal'
            ? e.clientX - containerRect.left
            : e.clientY - containerRect.top;

        // Calculate new size percentage
        let newSize = (currentPos / containerSize) * 100;

        // Apply constraints
        newSize = Math.max(minSize, Math.min(maxSize, newSize));

        setSize(newSize);
    }, [enabled, isDragging, containerRef, direction, minSize, maxSize]);

    // Handle mouse up
    const handleMouseUp = useCallback(() => {
        setIsDragging(false);
    }, []);

    // Add global mouse event listeners when dragging
    useEffect(() => {
        if (isDragging) {
            document.addEventListener('mousemove', handleMouseMove);
            document.addEventListener('mouseup', handleMouseUp);
            document.body.style.cursor = direction === 'horizontal' ? 'col-resize' : 'row-resize';
            document.body.style.userSelect = 'none';
        } else {
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
            document.body.style.cursor = 'auto';
            document.body.style.userSelect = 'auto';
        }

        return () => {
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
            document.body.style.cursor = 'auto';
            document.body.style.userSelect = 'auto';
        };
    }, [isDragging, handleMouseMove, handleMouseUp, direction]);

    return {
        size,
        isDragging,
        handleMouseDown
    };
}; 