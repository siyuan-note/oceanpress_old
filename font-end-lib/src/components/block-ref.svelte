<script lang="ts">
    /** ═════════🏳‍🌈 块引用 => 超链接 🏳‍🌈═════════  */
    export let src: string;
    export let title: string;
</script>

<style>
    [data-block-type="NodeBlockRefText"] {
        color: var(--b3-vditor-mark-block-color);
    }
</style>

<svelte:options tag="md2website-block-ref" /><a
    class="c-block-ref"
    data-block-type="NodeBlockRefText"
    href={src}>
    {title}
</a>
