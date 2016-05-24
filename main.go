package main

import (
        "flag"
	"strings"
        "fmt"
	"os"
	"bufio"
        "log"
        "net/http"

        "google.golang.org/api/googleapi/transport"
        "google.golang.org/api/youtube/v3"
)

var (
	seedFileName = flag.String("seed", "", "List of video id's and query strings (prefixed with ^) to generate contents from")
	developerKey = flag.String("developer-key", "", "Google API developer key")
        maxResults = flag.Int64("max-results", 50, "Max YouTube results")
	safeSearch = flag.String("safe-search", "none", "Set safe search (none (default), moderate, strict)")
	showChannels = flag.Bool("show-channels", false, "Show YouTube channels as well")
	showPlaylists = flag.Bool("show-playlists", false, "Show YouTube playlists as well")
	verbose = flag.Bool("verbose", false, "Enable verbose output")
)

func main() {
	flag.Parse()

	if *developerKey == "" {
		log.Fatalf("Developer key not specified. Obtain one from Google and pass to --developer-key.")
	}

	var seedFile = os.Stdin;

	if *seedFileName != "" {
		file, err := os.Open(*seedFileName)
		if err != nil {
			log.Fatalf("Error opening seed file: %v", err)
		}
		seedFile = file
	}

	client := &http.Client {
		Transport: &transport.APIKey{Key: *developerKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	scanner := bufio.NewScanner(seedFile)
	for scanner.Scan() {
		text := scanner.Text()
		// Make the API call to YouTube.
		call := service.Search.List("id,snippet").
			SafeSearch(*safeSearch).
			MaxResults(*maxResults)
		if strings.HasPrefix(text, "^") {
			call = call.Q(strings.TrimLeft(text, "^"))
		} else {
			call = call.RelatedToVideoId(text).Type("video")
		}
		response, err := call.Do()
		if err != nil {
			log.Printf("Error making search API call: %v", err)
		}

		// Group video, channel, and playlist results in separate lists.
		videos := make(map[string]string)
		channels := make(map[string]string)
		playlists := make(map[string]string)

		// Iterate through each item and add it to the correct list.
		for _, item := range response.Items {
			switch item.Id.Kind {
			case "youtube#video":
				videos[item.Id.VideoId] = item.Snippet.Title
			case "youtube#channel":
				channels[item.Id.ChannelId] = item.Snippet.Title
			case "youtube#playlist":
				playlists[item.Id.PlaylistId] = item.Snippet.Title
			}
		}

		printIDs("Videos", videos, *verbose)
		if *showChannels {
			printIDs("Channels", channels, *verbose)
		}
		if *showPlaylists {
			printIDs("Playlists", playlists, *verbose)
		}
	}
}

// Print the ID and title of each result in a list as well as a name that
// identifies the list. For example, print the word section name "Videos"
// above a list of video search results, followed by the video ID and title
// of each matching video.
func printIDs(sectionName string, matches map[string]string, verbose bool) {
	if verbose {
		fmt.Printf("%v:\n", sectionName)
	}
        for id, title := range matches {
		if verbose {
			fmt.Printf("[%v] %v\n", id, title)
		} else {
			fmt.Printf("%v\n", id)
		}
        }
	if verbose {
		fmt.Printf("\n\n")
	}
}
