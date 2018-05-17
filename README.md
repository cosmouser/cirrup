# cirrup

This is an early release of cirrup. Please forward any questions to
cosmo@ucsc.edu.

## What it Does
Cirrup is a web server that manages Jamf static groups defined by ldap attributes not captured by Jamf. The intended use is for automatically assigning computers belonging to users to groups based on attribute values from their ldap entry. For example, maintaining a static group of faculty and staff that acts dynamically like a smart group.

## Download
[Latest release](https://github.com/cosmouser/cirrup/releases)

## Install
1. Download the latest build from the links above
2. Create a cirrup user in the JSS with Static Computer Group update
privileges.
3. Create the JSS static groups that you want your ldap attribute values to map to
4. Create a ComputerInventoryCompleted webhook in the JSS that points to
the server that you'll be hosting cirrup off of. Cirrup listens on 8443
so point it to http://servername:8443/handle_cirrup
5. Fill out the provided config file and put it in the same directory as
the executable.
6. see usage

## usage

After filling out your config.toml, run cirrup without any arguments to initialize the caching database. Then, make sure to start cirrup with the -load flag to use the existing cache. If you run cirrup without the -load flag, it will truncate cache.db and recreate it.

