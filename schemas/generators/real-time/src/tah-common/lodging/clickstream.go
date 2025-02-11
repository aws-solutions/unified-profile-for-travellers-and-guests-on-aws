package lodging

// PRODUCT
var EVENT_ATTRIBUTE_ROOM_TYPE = "room_type"             //Room Type Code "DBL", "KING"
var EVENT_ATTRIBUTE_RATE_PLAN = "rate_plan"             //Rate Plan Code	"BAR"
var EVENT_ATTRIBUTE_CHECKIN_DATE = "checkin_date"       //checkin date
var EVENT_ATTRIBUTE_CHECKOUT_DATE = "checkout_date"     //checkout date
var EVENT_ATTRIBUTE_N_NIGHTS = "num_nights"             //Length of stay
var EVENT_ATTRIBUTE_HOTEL_CODE = "hotel_code"           //Hotel identifier
var EVENT_ATTRIBUTE_HOTEL_CODES = "hotel_code_list"     //list fo hotel identified (for multi avaialbility)
var EVENT_ATTRIBUTE_HOTEL_NAME = "hotel_name"           //human readable hotel name
var EVENT_ATTRIBUTE_PRODUCTS = "products"               //Products Price	["Room-Rate","Room-Rate","BKFST", "CMPGNE"]	Array
var EVENT_ATTRIBUTE_PRODUCTS_PRICES = "products_prices" //Products Price	["2103.00","2600.00","0.00","0.00","1200.00","1140.00","0.00","0.00","1200.00","1140.00","595.00","595.00","-805.00"]	Array
var EVENT_ATTRIBUTE_QUANTITIES = "quantities"           //Product Quantity	["3","3","3"]	Array
var EVENT_ATTRIBUTE_DESTINATION = "destination"         //destination / location
var EVENT_ATTRIBUTES_PICTURES = "product_pictures"      //product picture urls
var EVENT_ATTRIBUTES_PICTURE = "product_picture"        //product picture url

// GUESTS
var EVENT_ATTRIBUTE_NUM_GUEST_ADULTS = "num_guest_adults"     //Number of Adults	"2"
var EVENT_ATTRIBUTE_NUM_GUEST_CHILDREN = "num_guest_children" //Number of Children	"1"
var EVENT_ATTRIBUTE_TOTAL_GUESTS = "num_guests"               //Number of Passengers	"3"
