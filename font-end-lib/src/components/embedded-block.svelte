<script lang="ts">
    export let src: string;
    export let title: string;
</script>

<style>
    /** 嵌入块的样式 */
    .embedded_block {
        display: block;
        border: 1px solid var(--b3-border-color);
        border-radius: 4px;
        margin-bottom: 8px;
        box-sizing: border-box;
        white-space: initial;
    }
    .embedded_block-header {
        border-bottom: 1px solid var(--b3-border-color);
        padding: 8px 16px;
        background-color: var(--b3-theme-surface);
        display: flex;
        align-items: center;
    }
    .embedded_block-content {
        padding: 8px 16px;
    }
</style>

<svelte:options tag="oceanpress-embedded-block" />
<article class="embedded_block">
    <header class="embedded_block-header">
        <svg
            viewBox="0 0 16 16"
            version="1.1"
            width="16"
            height="16"
            style="margin-right:4px">
            <path
                fill-rule="evenodd"
                d="M4 9h1v1H4c-1.5 0-3-1.69-3-3.5S2.55 3 4 3h4c1.45 0 3 1.69 3 3.5 0 1.41-.91 2.72-2 3.25V8.59c.58-.45 1-1.27 1-2.09C10 5.22 8.98 4 8 4H4c-.98 0-2 1.22-2 2.5S3 9 4 9zm9-3h-1v1h1c1 0 2 1.22 2 2.5S13.98 12 13 12H9c-.98 0-2-1.22-2-2.5 0-.83.42-1.64 1-2.09V6.25c-1.09.53-2 1.84-2 3.25C6 11.31 7.55 13 9 13h4c1.45 0 3-1.69 3-3.5S14.5 6 13 6z" />
        </svg>
        <!-- {{template "blockRef" .}} -->
        <md2website-block-ref {src} {title} />
    </header>

    <div class="embedded_block-content">
        <slot />
    </div>
</article>
