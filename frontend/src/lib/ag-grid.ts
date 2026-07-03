/**
 * Shared AG Grid Configuration — Lazy Load
 *
 * AG Grid di-dynamic import hanya saat halaman yang membutuhkan dimuat.
 * Module registration dilakukan sekali via promise singleton.
 */

let agGridPromise: Promise<any> | null = null;

export async function getAgGrid() {
	if (agGridPromise) return agGridPromise;
	agGridPromise = (async () => {
		const mod = await import('ag-grid-community');

		const {
			ModuleRegistry,
			ClientSideRowModelModule,
			TextFilterModule,
			NumberFilterModule,
			DateFilterModule,
			ColumnAutoSizeModule,
			PaginationModule,
			RowSelectionModule,
			TextEditorModule,
			NumberEditorModule,
			DateEditorModule,
			SelectEditorModule,
			RowDragModule,
			QuickFilterModule,
			RowAutoHeightModule,
			TooltipModule,
			CsvExportModule,
			CellStyleModule,
		} = mod;

		// Register ONLY the modules we actually use
		ModuleRegistry.registerModules([
			ClientSideRowModelModule,
			TextFilterModule,
			NumberFilterModule,
			DateFilterModule,
			ColumnAutoSizeModule,
			PaginationModule,
			RowSelectionModule,
			TextEditorModule,
			NumberEditorModule,
			DateEditorModule,
			SelectEditorModule,
			RowDragModule,
			QuickFilterModule,
			RowAutoHeightModule,
			TooltipModule,
			CsvExportModule,
			CellStyleModule,
		]);

		return mod;
	})();
	return agGridPromise;
}
