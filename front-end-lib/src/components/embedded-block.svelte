<svelte:options tag="oceanpress-embedded-block" />

<script lang="ts">
  export let src: string;
  export let title: string;
</script>

<article class="embedded_block">
  <header class="embedded_block-header">
    <a href={src} title="打开 [{title}]({src})" style="display: flex;align-items: center;">
      <svg viewBox="0 0 32 32" style="height: 26px;width: 18px;fill: #fff">
        <path
          d="M24.038 7.962q3.305 0 5.634 2.366t2.329 5.671-2.329 5.671-5.634 2.366h-6.46v-3.080h6.46q2.028 0 3.493-1.465t1.465-3.493-1.465-3.493-3.493-1.465h-6.46v-3.080h6.46zM9.615 17.577v-3.155h12.77v3.155h-12.77zM3.005 16q0 2.028 1.465 3.493t3.493 1.465h6.46v3.080h-6.46q-3.305 0-5.634-2.366t-2.329-5.671 2.329-5.671 5.634-2.366h6.46v3.080h-6.46q-2.028 0-3.493 1.465t-1.465 3.493z"
        />
      </svg>
    </a>
  </header>

  <div class="embedded_block-content">
    <slot />
  </div>
</article>

<style>
  /** 嵌入块的样式 */
  .embedded_block {
    display: block;
    border: 1px solid var(--b3-border-color);
    border-radius: 4px;
    margin-bottom: 8px;
    box-sizing: border-box;
    white-space: initial;
    position: relative;
  }
  .embedded_block-header {
    position: absolute;
    right: 6px;
    top: 6px;
    height:22px;
    padding: 2px 4px;
    background-color: var(--b3-theme-on-surface);
    transition: var(--b3-transition);
    border-radius: 4px;
    display: none;
  }
  .embedded_block:hover .embedded_block-header{
    display: flex;
  }
  .embedded_block-content {
    padding: 8px 16px;
    background-color: var(--b3-theme-surface);
    border-left: 1px dashed var(--b3-border-color);
  }
</style>
