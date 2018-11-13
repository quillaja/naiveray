# Pre-parallel

## 01-1-2048.png
1920 x 1080 image, 2048 samples per px
Beginning render
Render complete. Writing to output.png
Took: 18054.49 s (5hr)
Sec/Sample: 0.00425 ms

## 03-1.png
800 x 600 image, 512 samples per px
Beginning render
Render complete. Writing to output.png
Took: 1252.54 s
Sec/Sample: 0.00510 ms

# Post-parallel

## 04-1.png
800 x 600 image, 512 samples per px
Beginning render
Render complete. Writing to output.png
Took: 857.30 s
Sec/Sample: 0.00349 ms
_Note: 31% speed up compared to 03-1png using 8 worker goroutines_

## unsaved
400 x 300 image, 32 samples per px
Beginning render
Render complete. Writing to output.png
Took: 8.39 s
Sec/Sample: 0.00218 ms

# Chunk size

## 640x480 32rays 
### 32x32 chunk (300 chunks)
1. Took: 26.40 s - Sec/Sample: 0.00269 ms
2. Took: 26.55 s - Sec/Sample: 0.00270 ms
3. Took: 28.59 s - Sec/Sample: 0.00291 ms
### 64x64 chunk (80 chunks)
1. Took: 25.40 s - Time/Sample: 0.00258 ms
2. Took: 26.92 s - Time/Sample: 0.00274 ms
3. Took: 29.93 s - Time/Sample: 0.00304 ms
### 128x128 chunk (20 chunks)
1. Took: 25.84 s - Time/Sample: 0.00263 ms
### 16x16 chunk (1200 chunks)
1. Took: 28.28 s - Time/Sample: 0.00288 ms