import os
import re
import glob

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # We want to replace the effect that destroys the grid and the effect that creates it.
    # Because Svelte 5 effects for agGrid are usually clustered together:
    
    # Let's find $effect(() => { ... gridApi.destroy(); ... });
    # and $effect(() => { ... agGridModule.createGrid ... });
    
    # Actually, a simpler way is to replace the items.length > 0 check!
    # Because items.length > 0 is the root cause of the empty grid not rendering.
    
    # Find: if (items.length > 0 && gridContainer
    # Replace: if (gridContainer
    
    # Wait, the other bug is that gridApi is not destroyed when showDetail becomes true!
    # Let's just find "gridApi.destroy();\n\t\t\tgridApi = null;" inside the first effect
    # and replace the condition.
    
    # Let's just write a regex that matches:
    # 	$effect(() => {
    # 		if (showForm && gridApi) {
    # 			gridApi.destroy();
    # 			gridApi = null;
    # 		}
    # 	});
    # and changes it to:
    # 	$effect(() => {
    # 		if (!gridContainer && gridApi) {
    # 			gridApi.destroy();
    # 			gridApi = null;
    # 		}
    # 	});
    
    new_content = re.sub(
        r'\$effect\(\(\) => \{\s*if \([^)]+\&\&\s*gridApi\) \{\s*gridApi\.destroy\(\);\s*gridApi = null;\s*\}\s*\}\);',
        r'$effect(() => {\n\t\tif (!gridContainer && gridApi) {\n\t\t\tgridApi.destroy();\n\t\t\tgridApi = null;\n\t\t}\n\t});',
        content
    )
    
    # Second effect replacement:
    # Remove "items.length > 0 && "
    new_content = new_content.replace('items.length > 0 && gridContainer', 'gridContainer')
    
    if new_content != content:
        with open(filepath, 'w') as f:
            f.write(new_content)
        print("Updated:", filepath)

for root, _, files in os.walk('frontend/src/routes/(app)'):
    for file in files:
        if file == '+page.svelte':
            process_file(os.path.join(root, file))
