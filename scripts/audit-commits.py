#!/usr/bin/env python3
"""Check strict single-file history, optionally after a supplied base commit."""
import subprocess,sys
revision=sys.argv[1]+'..HEAD' if len(sys.argv)>1 else 'HEAD'
commits=subprocess.check_output(['git','rev-list','--reverse',revision],text=True).splitlines()
bad=[]
for commit in commits:
 files=subprocess.check_output(['git','diff-tree','--root','--no-commit-id','--name-only','-r',commit],text=True).splitlines()
 if len(files)!=1:bad.append((commit,len(files)))
if bad:
 for commit,count in bad:print(f'{commit}: {count} files')
 sys.exit(1)
print(f'Strict ACDD: {len(commits)} single-file commits')
