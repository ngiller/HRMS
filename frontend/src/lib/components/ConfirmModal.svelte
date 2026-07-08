<script lang="ts">
	import AnimatedPresence from './AnimatedPresence.svelte';

  let {
    show = $bindable(true),
    title = 'Konfirmasi',
    message = '',
    confirmText = 'Ya, Hapus',
    confirmColor = 'red' as 'red' | 'blue' | 'green' | 'gray',
    isSaving = false,
    onConfirm = () => {},
    onCancel = () => {},
  }: {
    show?: boolean;
    title?: string;
    message?: string;
    confirmText?: string;
    confirmColor?: 'red' | 'blue' | 'green' | 'gray';
    isSaving?: boolean;
    onConfirm?: () => void;
    onCancel?: () => void;
  } = $props();

  const confirmColors: Record<string, string> = {
    red: 'bg-red-600 hover:bg-red-700 focus:ring-red-500',
    blue: 'bg-[#1A56DB] hover:bg-[#1e40af] focus:ring-[#1A56DB]',
    green: 'bg-green-600 hover:bg-green-700 focus:ring-green-500',
    gray: 'bg-gray-600 hover:bg-gray-700 focus:ring-gray-500',
  };
</script>

<AnimatedPresence show={show} type="scale" duration={200}>
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
	  class="fixed inset-0 flex items-center justify-center bg-black/50 backdrop-blur-sm z-50"
	  onclick={() => onCancel()}
	  onkeydown={(e) => { if (e.key === 'Escape') onCancel(); }}
	  role="dialog"
	  aria-modal="true"
	  tabindex="-1"
	>
	  <!-- svelte-ignore a11y_click_events_have_key_events -->
	  <div
	    onclick={(e) => e.stopPropagation()}
	    class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl max-w-sm w-full mx-4 p-6 transform transition-all"
	  >
	    <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">{title}</h2>
	    <p class="text-gray-700 dark:text-gray-300 mb-5">{message}</p>
	    <div class="flex justify-end gap-3">
	      <button
	        onclick={() => onCancel()}
	        disabled={isSaving}
	        class="px-4 py-2 rounded-lg text-sm font-medium border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition disabled:opacity-50 cursor-pointer"
	      >
	        Batal
	      </button>
	      <button
	        onclick={() => onConfirm()}
	        disabled={isSaving}
	        class="px-4 py-2 rounded-lg text-sm font-medium text-white transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer {confirmColors[confirmColor] || confirmColors.red}"
	      >
	        {#if isSaving}
	          <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
	        {/if}
	        {confirmText}
	      </button>
	    </div>
	  </div>
	</div>
</AnimatedPresence>
