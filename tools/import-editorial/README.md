# Henan Taihang editorial collection

This collection contains 23 notes, 10 guides and four clearly identified TuYou editor profiles. The posts are newly written destination information, not firsthand visitor reviews. They use the actual import date, and no historical likes, comments, registrations or trip participation are fabricated.

The 25 photos are from Wikimedia Commons. `data/photo-credits.json` records each source, author, license and SHA-256. The H5 repository publishes the corresponding images at `/editorial/<key>.jpg`; article photos link to a separate source and license page. Recheck attraction rules, transport and opening hours before publication because they can change.

From the backend repository root, validate without writing:

```sh
go run ./tools/import-editorial \
  --manifest tools/import-editorial/data/content.json \
  --credits tools/import-editorial/data/photo-credits.json \
  --photo-base https://tuyou.evenstar.love/editorial/
```

After backing up production MySQL and verifying the photos are online, add `--config <private-server-config.yaml> --apply`. The importer checks the published image hashes, then inserts authors, posts, tags and image links in one database transaction. It leaves existing records alone and refuses title collisions or changed editorial accounts. Run it only after the H5 assets are deployed. Never commit the private server config.
