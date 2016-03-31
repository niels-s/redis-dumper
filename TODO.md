1. process multiple db's at once
2. split files into chuck when they get to large
3. optionally get ttl (now restore uses 0 which means no ttl)
4. split looping keys and appending into seperate go routine
