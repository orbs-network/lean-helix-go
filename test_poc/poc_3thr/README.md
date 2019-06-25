# 3-threaded model POC

## Goroutines

* Mainloop
* Termloop
* Viewloop (long-running ops are synchronous within it)

We consider this the best solution engineering-wise, but it is the costliest so not developing it for now.
