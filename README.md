# GCloudz

Tool for zipping files stored in Google Cloud Storage, just point and shoot like so;
```
zipper, err := gcloudz.NewWithBucketNamed(ctx, "mybucket.mysite.com")
zipper.Zip("MyFolder", "MyFolder.zip", "application/zip", metadata);
```

An extended example;
```
import (
	"github.com/ranisputnik/gcloudz"
	"google.golang.org/appengine"
	"net/http"
)

...

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	zipper, err := gcloudz.NewWithBucketNamed(ctx, "mybucket.mysite.com")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := zipper.Zip("TheGame", "TheGame.love", "application/x-love-game", nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
```

If you aren't working in App Engine, then you can load your Google Cloud credentials from a file. Create a service account on "Credentials" for your project at https://console.developers.google.com to download a JSON key file.
```
myCredentialsFile := "path/to/credentials.json"
gcloudz.NewWithCredentials(ctx, myBucketName, myCredentialsFiles)
```

Or, if you need a specially configured client, provide a bucket handle yourself;
```
import (
  "github.com/ranisputnik/gcloudz"
  "google.golang.org/cloud/storage"
)

...

// TODO create the bucket - see cloud storage docs
gcloudz.NewWithBucket(ctx, bucket)
```

That's it, enjoy.
