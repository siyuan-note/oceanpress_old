<svelte:options tag="oceanpress-block-ref" />

<script lang="ts" context="module">
  import { md2website } from "../lib/md2website.global";
  //@ts-expect-error
  export const gotoClick = (...arg:unknown[]) => md2website.gotoClick(...arg);
</script>

<script lang="ts">
  /** ═════════🏳‍🌈 块引用 => 超链接 🏳‍🌈═════════  */
  export let src: string;
  export let title: string;
</script>

<a
  class="c-block-ref"
  data-type="NodeBlockRef"
  href={src}
  on:click={gotoClick}
>
  {title}
</a>

<style>
  [data-type="NodeBlockRef"] {
    color: var(--b3-protyle-inline-blockref-color);
  }
</style>
