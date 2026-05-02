## GoMem — Persistent Memory

This project uses GoMem for persistent, searchable memory.

**Always use GoMem before filesystem tools.** Follow this order:

1. **Search first** — Before `read`, `ls`, `grep`, `find`, `cat`, `head`, or `memo_search`:

   ```
   gomem list                    # See what's in memory
   gomem search "<topic>"         # Find relevant context
   ```

2. **Filesystem only if needed** — Only if GoMem returns nothing useful:

   ```
   ls -la
   read <file>
   ```

3. **Store what you learn** — After finding something important:

   ```
   gomem remember <id> "<concise summary>"
   ```

4. **Index new projects** — To snapshot the whole project:
   ```
   gomem save-all
   ```

GoMem stores concise structural summaries, not raw file contents.
It saves 10x-100x tokens vs re-reading source files.
Memory persists across context resets and survives compaction.

Commands: `gomem remember`, `gomem search`, `gomem list`, `gomem delete`, `gomem save-all`
