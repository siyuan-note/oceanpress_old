<svelte:options tag="md2website-block-ref" />

<script lang="ts" context="module">
  import { md2website } from "../lib/md2website.global";
  //@ts-expect-error
  export const gotoClick = (...arg) => md2website.gotoClick(...arg);
</script>

<script lang="ts">
  /** ═════════🏳‍🌈 块引用 => 超链接 🏳‍🌈═════════  */
  export let src: string;
  export let title: string;
</script>

<a
  class="c-block-ref"
  data-type="NodeBlockRefText"
  href={src}
  on:click={gotoClick}
>
  {title}
</a>

<style>
  [data-type="NodeBlockRefText"] {
    color: var(--b3-protyle-inline-blockref-color);
  }
</style>
