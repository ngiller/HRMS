<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let open: boolean = false;
  export let title: string = 'Konfirmasi';
  export let message: string = '';

  const dispatch = createEventDispatcher();

  function confirm() {
    dispatch('confirm');
    close();
  }
  function cancel() {
    dispatch('cancel');
    close();
  }
  function close() {
    open = false;
  }
</script>

{#if open}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 flex items-center justify-center bg-black/30 backdrop-blur-sm z-50"
    on:click|self={cancel}
    on:keydown={(e) => { if (e.key === 'Escape') cancel(); }}
    role="dialog"
    aria-modal="true"
    tabindex="-1"
  >
    <div
      class="bg-white dark:bg-gray-800 rounded-xl shadow-xl max-w-sm w-full mx-4 p-6 transform transition-all
             motion-safe:animate-fade-in motion-safe:animate-scale-in"
    >
      <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">{title}</h2>
      <p class="text-gray-700 dark:text-gray-300 mb-5">{message}</p>
      <div class="flex justify-end gap-3">
        <button
          class="px-4 py-2 rounded-lg text-sm font-medium bg-gray-200 dark:bg-gray-700 text-gray-800
                 dark:text-gray-200 hover:bg-gray-300 dark:hover:bg-gray-600 transition"
          on:click={cancel}
        >
          Batal
        </button>
        <button
          class="px-4 py-2 rounded-lg text-sm font-medium bg-red-600 text-white hover:bg-red-700 transition"
          on:click={confirm}
        >
          Ya, Hapus
        </button>
      </div>
    </div>
  </div>
{/if}
