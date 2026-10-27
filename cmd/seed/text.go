package main

import (
	"fmt"
	"math/rand"
	"strings"
)

var firstNames = []string{
	"Aarav", "Vivaan", "Aditya", "Vihaan", "Arjun", "Sai", "Reyansh", "Ayaan", "Krishna", "Ishaan",
	"Shaurya", "Atharv", "Advik", "Pranav", "Kabir", "Rohan", "Karan", "Rahul", "Nikhil", "Varun",
	"Siddharth", "Manish", "Amit", "Rajesh", "Suresh", "Vikram", "Anil", "Deepak", "Harsh", "Yash",
	"Aadhya", "Ananya", "Diya", "Saanvi", "Pari", "Anika", "Navya", "Myra", "Sara", "Ira",
	"Aisha", "Riya", "Kavya", "Meera", "Neha", "Pooja", "Priya", "Sneha", "Shreya", "Tanvi",
	"Isha", "Nisha", "Divya", "Anjali", "Kriti", "Lakshmi", "Swati", "Payal", "Ritika", "Zoya",
	"Farhan", "Imran", "Sameer", "Arnav", "Dev", "Om", "Parth", "Tejas", "Gaurav", "Mohit",
	"Fatima", "Sana", "Mehak", "Simran", "Jasleen", "Harleen", "Gurpreet", "Manpreet", "Aditi", "Bhavna",
}

var lastNames = []string{
	"Sharma", "Verma", "Gupta", "Singh", "Kumar", "Patel", "Shah", "Mehta", "Joshi", "Desai",
	"Iyer", "Iyengar", "Nair", "Menon", "Pillai", "Reddy", "Rao", "Naidu", "Chowdhury", "Banerjee",
	"Mukherjee", "Chatterjee", "Das", "Bose", "Sen", "Ghosh", "Kapoor", "Khanna", "Malhotra", "Arora",
	"Bhatia", "Chopra", "Saxena", "Agarwal", "Jain", "Kulkarni", "Deshpande", "Patil", "Pawar", "Shinde",
	"Thakur", "Yadav", "Mishra", "Tiwari", "Pandey", "Dubey", "Khan", "Ahmed", "Sheikh", "Fernandes",
	"D'Souza", "Pinto", "Gill", "Sandhu", "Grewal", "Bajwa", "Hegde", "Shetty", "Kamath", "Bhat",
}

var reviewTitles = map[int][]string{
	5: {"Absolutely love it", "Excellent quality", "Worth every rupee", "Exceeded my expectations", "Highly recommended",
		"Perfect!", "Best purchase this year", "Superb product", "Five stars, no doubt", "Brilliant value for money"},
	4: {"Very good product", "Happy with the purchase", "Good value", "Does the job well", "Nice quality",
		"Pretty good overall", "Solid buy", "Recommended", "Good, with minor niggles", "Satisfied"},
	3: {"Average", "It's okay", "Decent for the price", "Not bad, not great", "Mixed feelings",
		"Does what it says", "Could be better", "Fair enough"},
	2: {"Disappointed", "Not as described", "Below expectations", "Quality could be better", "Wouldn't buy again"},
	1: {"Very poor", "Waste of money", "Stopped working quickly", "Do not buy", "Terrible experience"},
}

var reviewOpeners = map[int][]string{
	5: {"I have been using this %s for a few weeks now and it is fantastic.", "Really impressed with this %s.",
		"This %s is exactly what I was looking for.", "Bought this %s after reading the reviews and they were right."},
	4: {"Good %s overall.", "I'm happy with this %s.", "The %s works well for my needs.", "Decent %s for the price."},
	3: {"The %s is okay.", "This %s is average at best.", "Mixed experience with this %s."},
	2: {"Not very happy with this %s.", "The %s did not live up to the description.", "Expected more from this %s."},
	1: {"Really unhappy with this %s.", "The %s was a letdown.", "I regret buying this %s."},
}

var reviewDetails = map[int][]string{
	5: {"Build quality is excellent and it looks premium.", "Delivery was quick and the packaging was secure.",
		"Great value compared to similar products.", "My family loves it too.", "Would definitely buy again.",
		"Works perfectly, no complaints at all."},
	4: {"Quality is good, though the packaging could be better.", "Delivery took a day longer than promised.",
		"Slightly pricey but worth it.", "Does everything I need.", "Would recommend to friends."},
	3: {"It works, but feels a bit cheap.", "Took a while to arrive.", "Good enough for occasional use.",
		"The colour is slightly different from the photos."},
	2: {"The finish is not great.", "It started showing problems after a week.", "Customer support was slow to respond."},
	1: {"It arrived damaged.", "It stopped working within days.", "The quality is nowhere near the photos."},
}

func pick[T any](rng *rand.Rand, items []T) T {
	return items[rng.Intn(len(items))]
}

// reviewText writes a short review for a product type, e.g. "smartphone".
func reviewText(rng *rand.Rand, rating int, noun string) (string, string) {
	title := pick(rng, reviewTitles[rating])
	parts := []string{fmt.Sprintf(pick(rng, reviewOpeners[rating]), noun)}

	details := append([]string{}, reviewDetails[rating]...)
	rng.Shuffle(len(details), func(i, j int) { details[i], details[j] = details[j], details[i] })
	parts = append(parts, details[:1+rng.Intn(2)]...)

	return title, strings.Join(parts, " ")
}

// productNoun names what a department sells, for review text.
var productNoun = map[string]string{
	"beauty": "product", "fragrances": "fragrance", "furniture": "piece of furniture", "groceries": "product",
	"home-decoration": "decor piece", "kitchen-accessories": "kitchen tool", "laptops": "laptop",
	"mens-shirts": "shirt", "mens-shoes": "pair of shoes", "mens-watches": "watch", "mobile-accessories": "accessory",
	"motorcycle": "bike", "skin-care": "product", "smartphones": "phone", "sports-accessories": "gear",
	"sunglasses": "pair of sunglasses", "tablets": "tablet", "tops": "top", "vehicle": "car", "womens-bags": "bag",
	"womens-dresses": "dress", "womens-jewellery": "piece", "womens-shoes": "pair of shoes", "womens-watches": "watch",
}
