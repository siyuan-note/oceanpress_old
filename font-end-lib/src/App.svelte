<script lang="ts">
	import "./all_components"
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
<div class="row" title="嵌入块组件">
	<md2website-embedded-block src="11" title="22">
		33
	</md2website-embedded-block>
</div>
