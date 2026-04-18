import { onMounted, onUnmounted, type Ref } from 'vue';

export function useClickOutside(
    targetRef: Ref<HTMLElement | null>, 
    callback: () => void
) {
    const handleClick = (event: MouseEvent) => {
        if (targetRef.value && !targetRef.value.contains(event.target as Node)) {
            callback();
        }
    };

    onMounted(() => document.addEventListener('mousedown', handleClick));
    onUnmounted(() => document.removeEventListener('mousedown', handleClick));
}