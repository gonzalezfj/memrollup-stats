import time
import array

# Allocate ~100MB of memory
mem = array.array('B', [0] * (100 * 1024 * 1024))

for i in range(5):
    time.sleep(1)
    print(f"Iteration {i}")
    mem = array.array('B', [0] * (100 * 1024 * 1024)) # 100MB