# vgmusic.go

(unofficial) Go API for vgmusic.com. This project is in no way affiliated with or sponsered by Mike Newman or any of the staff at VGMusic.

## Bruh you already made a Python version!11!1!

I rewrote vgmusic.py in Go because I wanted to learn it (this is probably the fastest way for me to learn a new programming language).

## Improvements over vgmusic.py

- Simplified object model. Songs are now stored in a flat map by their MD5 checksums, which should make searching easier.
- Parallel processing goodness. Channels (goroutines) can be utilised to speed up parsing/searching even more than Python's multiprocessing (no GIL).
  Plus, they're lightweight.
- New-files parsing. Instead of just parsing the main archive, vgmusic.go can parse new files that have not been moved to the main archive yet.

Also, the command-line `vgmusic` tool will still be available.
