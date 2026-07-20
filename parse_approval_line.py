import pandas as pd
df = pd.read_excel('employee_template.xlsx')
approval_cols = [col for col in df.columns if 'approval' in col.lower()]
print("Approval related columns:", approval_cols)
if approval_cols:
    for col in approval_cols:
        print(f"\nUnique values in {col}:")
        print(df[col].dropna().unique())
else:
    # Print the 39th column just in case
    col_name = df.columns[39]
    print(f"\nColumn 39 is: {col_name}")
    print(df[col_name].dropna().unique())
