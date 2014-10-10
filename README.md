# CloudFlare Configure

A utility to configure [CloudFlare CDN] settings from version-controllable
JSON files and using [CloudFlare's v4 API].

[CloudFlare CDN]: https://www.cloudflare.com/features-cdn
[CloudFlare's v4 API]: http://developers.cloudflare.com/next/

## Compiling

You will need [Go]. Any 1.x version should do. To compile the binary:

    go build

[Go]: http://golang.org/

## Usage

List the available zones:

    ➜  cdn-configs git:(master) ✗ ./cloudflare-configure -email user@example.com -key 123 -list-zones
    4986183da7c16aab483d31ac6bb4cb7b    foo.example.com
    15f14360e93a76824ab7d49a4533d970    bar.example.com
    d1082145f48bb35a023c6ec3a7897837    baz.example.com

Download the current configuration for a zone:

    ➜  cdn-configs git:(master) ✗ ./cloudflare-configure -email user@example.com -key 123 -zone 456 -file myzone.json -download
    2014/10/08 09:57:48 Saving configuration..

Modify a setting:

    ➜  cdn-configs git:(master) ✗ gsed -ri 's/("ipv6": )"off"/\1"on"/' myzone.json

Preview the changes:

    ➜  cdn-configs git:(master) ✗ ./cloudflare-configure -email user@example.com -key 123 -zone 456 -file myzone.json -dry-run
    2014/10/08 10:11:22 Dry run mode. Changes won't be submitted
    2014/10/08 10:11:22 Comparing and updating configuration..
    2014/10/08 10:11:22 Changing setting: ipv6 off -> on

Upload the changes:

    ➜  cdn-configs git:(master) ✗ ./cloudflare-configure -email user@example.com -key 123 -zone 456 -file myzone.json
    2014/10/08 10:11:58 Comparing and updating configuration..
    2014/10/08 10:11:58 Changing setting: ipv6 off -> on

Use the `-help` argument to see all of the options available.

## Considerations

The following caveats and limitations should be borne in mind:

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
