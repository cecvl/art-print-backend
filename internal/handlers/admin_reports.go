package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
)

// SalesMonthlyHandler returns sales aggregated by month
// Query params: from (ISO), to (ISO), shopId, artistId
func SalesMonthlyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	qFrom := r.URL.Query().Get("from")
	qTo := r.URL.Query().Get("to")
	shopId := r.URL.Query().Get("shopId")
	artistId := r.URL.Query().Get("artistId")

	var from time.Time
	var to time.Time
	var err error
	if qTo == "" {
		to = time.Now()
	} else {
		to, err = time.Parse(time.RFC3339, qTo)
		if err != nil {
			http.Error(w, "invalid to", http.StatusBadRequest)
			return
		}
	}
	if qFrom == "" {
		from = to.AddDate(0, -12, 0)
	} else {
		from, err = time.Parse(time.RFC3339, qFrom)
		if err != nil {
			http.Error(w, "invalid from", http.StatusBadRequest)
			return
		}
	}

	// Query orders in range
	q := firebase.FirestoreClient.Collection("orders").Where("createdAt", ">=", from).Where("createdAt", "<=", to).OrderBy("createdAt", firestore.Asc)
	if shopId != "" {
		q = q.Where("printShopId", "==", shopId)
	}
	if artistId != "" {
		// We'll include orders that contain items for that artist by scanning; Firestore cannot query nested arrays for artistId so we'll filter client-side
	}

	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("❌ failed to query orders for sales report: %v", err)
		http.Error(w, "failed to query orders", http.StatusInternalServerError)
		return
	}

	// aggregate by month
	series := map[string]map[string]float64{} // month -> { orders: count, revenue: value }
	counts := map[string]int{}

	for _, d := range docs {
		var o models.Order
		if err := d.DataTo(&o); err != nil {
			continue
		}
		// Optionally filter by artistId: if provided, ensure any CartItem.ArtworkID belongs to artist
		if artistId != "" {
			// naive: skip if no matching artist in order items (we don't have artwork->artist cached)
			matchesArtist := false
			for _, it := range o.Items {
				// try to fetch artwork and compare artistId
				artDoc, err := firebase.FirestoreClient.Collection("artworks").Doc(it.ArtworkID).Get(ctx)
				if err != nil {
					continue
				}
				var art models.Artwork
				if err := artDoc.DataTo(&art); err == nil {
					if art.ArtistID == artistId {
						matchesArtist = true
						break
					}
				}
			}
			if !matchesArtist {
				continue
			}
		}

		// only count confirmed/completed orders
		if o.Status != "confirmed" && o.Status != "completed" {
			continue
		}
		month := o.CreatedAt.Format("2006-01")
		if _, ok := series[month]; !ok {
			series[month] = map[string]float64{"revenue": 0}
			counts[month] = 0
		}
		series[month]["revenue"] += o.TotalAmount
		counts[month]++
	}

	// build ordered result
	months := make([]string, 0, len(series))
	for m := range series {
		months = append(months, m)
	}
	// sort months ascending
	// use time parse
	sortKeys := func(a []string) {
		// simple bubble sort small N
		for i := 0; i < len(a); i++ {
			for j := i + 1; j < len(a); j++ {
				if a[i] > a[j] {
					a[i], a[j] = a[j], a[i]
				}
			}
		}
	}
	sortKeys(months)

	out := make([]map[string]interface{}, 0, len(months))
	for _, m := range months {
		out = append(out, map[string]interface{}{"month": m, "orders": counts[m], "revenue": series[m]["revenue"]})
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"series": out})
}
