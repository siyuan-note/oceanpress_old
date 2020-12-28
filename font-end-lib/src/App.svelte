<script lang="ts">
	import Test from "./components/test.svelte";
	import Md from "./components/md.svelte";
	export let md: any;
</script>

<style>
	.row {
		border: 1px solid #333;
	}
</style>

<svelte:options tag="md2website-app" />
<h1>全部 web-components 组件的展示</h1>
<div class="row" title="测试组件">
	<md2website-test>444</md2website-test>
</div>
<div class="row" title="md渲染组件">
	<md2website-md bind:this={md}>444</md2website-md>
</div>
