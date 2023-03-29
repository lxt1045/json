package testdata

type Twitter struct {
	StatusesST       []StatusesST     `json:"statuses"`
	SearchMetadataST SearchMetadataST `json:"search_metadata"`
}
type MetadataST struct {
	ResultType      string `json:"result_type"`
	IsoLanguageCode string `json:"iso_language_code"`
}
type Description struct {
	Urls []interface{} `json:"urls"`
}
type EntitiesST struct {
	Description  Description    `json:"description"`
	HashtagsST   []interface{}  `json:"hashtags"`
	Symbols      []interface{}  `json:"symbols"`
	Urls         []interface{}  `json:"urls"`
	UserMentions []UserMentions `json:"user_mentions"`
	Media        []Media        `json:"media"`
}
type UserST struct {
	ID                             int         `json:"id"`
	IDStr                          string      `json:"id_str"`
	Name                           string      `json:"name"`
	ScreenName                     string      `json:"screen_name"`
	Location                       string      `json:"location"`
	Description                    string      `json:"description"`
	URL                            interface{} `json:"url"`
	EntitiesST                     EntitiesST  `json:"entities"`
	Protected                      bool        `json:"protected"`
	FollowersCount                 int         `json:"followers_count"`
	FriendsCount                   int         `json:"friends_count"`
	ListedCount                    int         `json:"listed_count"`
	CreatedAt                      string      `json:"created_at"`
	FavouritesCount                int         `json:"favourites_count"`
	UtcOffset                      int         `json:"utc_offset"`
	TimeZone                       string      `json:"time_zone"`
	GeoEnabled                     bool        `json:"geo_enabled"`
	Verified                       bool        `json:"verified"`
	StatusesCount                  int         `json:"statuses_count"`
	Lang                           string      `json:"lang"`
	ContributorsEnabled            bool        `json:"contributors_enabled"`
	IsTranslator                   bool        `json:"is_translator"`
	IsTranslationEnabled           bool        `json:"is_translation_enabled"`
	ProfileBackgroundColor         string      `json:"profile_background_color"`
	ProfileBackgroundImageURL      string      `json:"profile_background_image_url"`
	ProfileBackgroundImageURLHTTPS string      `json:"profile_background_image_url_https"`
	ProfileBackgroundTile          bool        `json:"profile_background_tile"`
	ProfileImageURL                string      `json:"profile_image_url"`
	ProfileImageURLHTTPS           string      `json:"profile_image_url_https"`
	ProfileBannerURL               string      `json:"profile_banner_url"`
	ProfileLinkColor               string      `json:"profile_link_color"`
	ProfileSidebarBorderColor      string      `json:"profile_sidebar_border_color"`
	ProfileSidebarFillColor        string      `json:"profile_sidebar_fill_color"`
	ProfileTextColor               string      `json:"profile_text_color"`
	ProfileUseBackgroundImage      bool        `json:"profile_use_background_image"`
	DefaultProfile                 bool        `json:"default_profile"`
	DefaultProfileImage            bool        `json:"default_profile_image"`
	Following                      bool        `json:"following"`
	FollowRequestSent              bool        `json:"follow_request_sent"`
	Notifications                  bool        `json:"notifications"`
}
type UserMentions struct {
	ScreenName string `json:"screen_name"`
	Name       string `json:"name"`
	ID         int    `json:"id"`
	IDStr      string `json:"id_str"`
	Indices    []int  `json:"indices"`
}
type Medium struct {
	W      int    `json:"w"`
	H      int    `json:"h"`
	Resize string `json:"resize"`
}
type Small struct {
	W      int    `json:"w"`
	H      int    `json:"h"`
	Resize string `json:"resize"`
}
type Thumb struct {
	W      int    `json:"w"`
	H      int    `json:"h"`
	Resize string `json:"resize"`
}
type Large struct {
	W      int    `json:"w"`
	H      int    `json:"h"`
	Resize string `json:"resize"`
}
type Sizes struct {
	Medium Medium `json:"medium"`
	Small  Small  `json:"small"`
	Thumb  Thumb  `json:"thumb"`
	Large  Large  `json:"large"`
}
type Media struct {
	ID                int64  `json:"id"`
	IDStr             string `json:"id_str"`
	Indices           []int  `json:"indices"`
	MediaURL          string `json:"media_url"`
	MediaURLHTTPS     string `json:"media_url_https"`
	URL               string `json:"url"`
	DisplayURL        string `json:"display_url"`
	ExpandedURL       string `json:"expanded_url"`
	Type              string `json:"type"`
	Sizes             Sizes  `json:"sizes"`
	SourceStatusID    int64  `json:"source_status_id"`
	SourceStatusIDStr string `json:"source_status_id_str"`
}
type RetweetedStatus struct {
	MetadataST           MetadataST  `json:"metadata"`
	CreatedAt            string      `json:"created_at"`
	ID                   int64       `json:"id"`
	IDStr                string      `json:"id_str"`
	Text                 string      `json:"text"`
	Source               string      `json:"source"`
	Truncated            bool        `json:"truncated"`
	InReplyToStatusID    interface{} `json:"in_reply_to_status_id"`
	InReplyToStatusIDStr interface{} `json:"in_reply_to_status_id_str"`
	InReplyToUserID      interface{} `json:"in_reply_to_user_id"`
	InReplyToUserIDStr   interface{} `json:"in_reply_to_user_id_str"`
	InReplyToScreenName  interface{} `json:"in_reply_to_screen_name"`
	UserST               UserST      `json:"user"`
	Geo                  interface{} `json:"geo"`
	Coordinates          interface{} `json:"coordinates"`
	Place                interface{} `json:"place"`
	Contributors         interface{} `json:"contributors"`
	RetweetCount         int         `json:"retweet_count"`
	FavoriteCount        int         `json:"favorite_count"`
	EntitiesST           EntitiesST  `json:"entities"`
	Favorited            bool        `json:"favorited"`
	Retweeted            bool        `json:"retweeted"`
	PossiblySensitive    bool        `json:"possibly_sensitive"`
	Lang                 string      `json:"lang"`
}

type HashtagsST struct {
	Text    string `json:"text"`
	Indices []int  `json:"indices"`
}
type StatusesST struct {
	MetadataST           MetadataST      `json:"metadata"`
	CreatedAt            string          `json:"created_at"`
	ID                   int64           `json:"id"`
	IDStr                string          `json:"id_str"`
	Text                 string          `json:"text"`
	Source               string          `json:"source"`
	Truncated            bool            `json:"truncated"`
	InReplyToStatusID    interface{}     `json:"in_reply_to_status_id"`
	InReplyToStatusIDStr interface{}     `json:"in_reply_to_status_id_str"`
	InReplyToUserID      int             `json:"in_reply_to_user_id"`
	InReplyToUserIDStr   string          `json:"in_reply_to_user_id_str"`
	InReplyToScreenName  string          `json:"in_reply_to_screen_name"`
	UserST               UserST          `json:"user,omitempty"`
	Geo                  interface{}     `json:"geo"`
	Coordinates          interface{}     `json:"coordinates"`
	Place                interface{}     `json:"place"`
	Contributors         interface{}     `json:"contributors"`
	RetweetCount         int             `json:"retweet_count"`
	FavoriteCount        int             `json:"favorite_count"`
	EntitiesST           EntitiesST      `json:"entities,omitempty"`
	Favorited            bool            `json:"favorited"`
	Retweeted            bool            `json:"retweeted"`
	Lang                 string          `json:"lang"`
	RetweetedStatus      RetweetedStatus `json:"retweeted_status,omitempty"`
	PossiblySensitive    bool            `json:"possibly_sensitive,omitempty"`
}
type SearchMetadataST struct {
	CompletedIn float64 `json:"completed_in"`
	MaxID       int64   `json:"max_id"`
	MaxIDStr    string  `json:"max_id_str"`
	NextResults string  `json:"next_results"`
	Query       string  `json:"query"`
	RefreshURL  string  `json:"refresh_url"`
	Count       int     `json:"count"`
	SinceID     int     `json:"since_id"`
	SinceIDStr  string  `json:"since_id_str"`
}
