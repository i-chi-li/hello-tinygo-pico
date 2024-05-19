build:
	$(MAKE) -C blink_led build --no-print-directory
	$(MAKE) -C hello build --no-print-directory
	$(MAKE) -C uart build --no-print-directory

