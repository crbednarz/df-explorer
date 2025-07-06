.PHONY: all build clean

# Directories and file names
LIBVTERM_DIR := libvterm
LIBVTERM_LIB := $(LIBVTERM_DIR)/.libs/libvterm.a
GO_PROJECT := df-explorer

all: build

build: $(LIBVTERM_LIB)
	@echo "Building Go project..."
	go build -o $(GO_PROJECT)

$(LIBVTERM_LIB):
	@echo "Building libvterm..."
	$(MAKE) -C $(LIBVTERM_DIR)

clean:
	@echo "Cleaning libvterm..."
	$(MAKE) -C $(LIBVTERM_DIR) clean
	@echo "Cleaning Go project..."
	rm -f $(GO_PROJECT)
