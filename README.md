# CloudFlare Configure

A utility to configure [CloudFlare CDN] settings from version-controllable
JSON files and using [CloudFlare's v4 API].

[CloudFlare CDN]: https://www.cloudflare.com/features-cdn
[CloudFlare's v4 API]: http://developers.cloudflare.com/next/

## Compiling

You will need [Go]. Any 1.x version should do. To compile the binary:

    make

[Go]: http://golang.org/

## Usage

Set some environment variables. These are *not* required, but they simplify
our examples and prevent the key from being stored in your shell history:

    ➜  cdn-configs git:(master) export CF_EMAIL=user@example.com
    ➜  cdn-configs git:(master)  export CF_KEY=b58996c504c5638798eb6b511e6f49af

List the available zones:

    ➜  cdn-configs git:(master) ./cloudflare-configure --email ${CF_EMAIL} --key ${CF_KEY} zones
    4986183da7c16aab483d31ac6bb4cb7b    foo.example.com
    15f14360e93a76824ab7d49a4533d970    bar.example.com
    d1082145f48bb35a023c6ec3a7897837    baz.example.com

Download the current configuration for a zone:

    ➜  cdn-configs git:(master) ./cloudflare-configure --email ${CF_EMAIL} --key ${CF_KEY} download 4986183da7c16aab483d31ac6bb4cb7b myzone.json
    2014/10/17 14:01:54 Saving config to: myzone.json

Modify some settings:

    ➜  cdn-configs git:(master) gsed -ri 's/("ipv6": )"off"/\1"on"/' myzone.json
    ➜  cdn-configs git:(master) gsed -ri 's/("browser_cache_ttl": )14400/\17200/' myzone.json

Review the changes:

    ➜  cdn-configs git:(master) ./cloudflare-configure --email ${CF_EMAIL} --key ${CF_KEY} upload 4986183da7c16aab483d31ac6bb4cb7b myzone.json --dry-run
    2014/10/17 14:16:06 Would have changed setting "ipv6" from "off" to "on"
    2014/10/17 14:16:06 Would have changed setting "browser_cache_ttl" from 14400 to 7200

Upload the changes:

    ➜  cdn-configs git:(master) ./cloudflare-configure --email ${CF_EMAIL} --key ${CF_KEY} upload 4986183da7c16aab483d31ac6bb4cb7b myzone.json
    2014/10/17 14:16:15 Changing setting "ipv6" from "off" to "on"
    2014/10/17 14:16:16 Changing setting "browser_cache_ttl" from 14400 to 7200

Use the `--help` argument to see all of the sub-commands and flags available.

## Considerations

The following caveats and limitations should be borne in mind:

- It can't manage "page rules", which are used to configure protocol
  redirects or caching of all content types, because they aren't currently
  supported by the API.
- If the key names in the local and remote configurations differ, for
  example if you have made a typo or CloudFlare introduce a new feature,
  then a message will be logged and you will need to update your
  configuration manually (compare with `-download`).
- It is assumed that any keys that need modifying have API endpoints of the
  same name, eg. `{"id":"always_online",…}` can be written at
  `/v4/zones/123/settings/always_online`. This appears to hold true.
- It doesn't respect `{"editable":false,…}"` which CloudFlare use to
  indicate that a key is immutable. Such keys probably don't have write
  endpoints (see the last point) and will fail, but they should also never
  change in your config.
- It will abort and cease to make any other changes as soon as it encounters
  an error.
