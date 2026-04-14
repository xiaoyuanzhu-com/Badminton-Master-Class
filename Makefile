.PHONY: build validate upload clean

# Validate content files against schemas
validate:
	python3 data/content/validate.py

# Validate + compile bmc.db + copy to app bundles
build:
	python3 data/build.py

# Validate + compile + copy + upload to OSS
upload:
	python3 data/build.py --upload

# Remove compiled database
clean:
	rm -f data/bmc.db
