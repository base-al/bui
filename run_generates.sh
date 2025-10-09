#!/bin/bash

cd "$(dirname "$0")"

echo "Running all module generations..."

./bui generate region designation:translation sort_order:int parent:belongsTo:Region country:belongsTo:Country

./bui generate user_role name:string code:string permissions:json

./bui generate attribute_group name:translation code:string sort_order:int

./bui generate attribute attribute_group:belongsTo:AttributeGroup code:string label:translation input_type:string is_required:bool is_unique:bool is_filterable:bool is_searchable:bool validation_rules:string sort_order:int

./bui generate attribute_option attribute:belongsTo:Attribute value:string label:translation sort_order:int

./bui generate attribute_value entity_type:string entity_id:int attribute:belongsTo:Attribute value_text:text value_decimal:float value_int:int value_datetime:datetime

./bui generate product_category name:translation slug:string parent:belongsTo:ProductCategory sort_order:int is_active:bool

./bui generate classification name:translation code:string

./bui generate filling name:translation volume:float unit:string

./bui generate trading_unit name:string code:string conversion_factor:int

./bui generate producer name:string slug:string description:text country:belongsTo:Country region:belongsTo:Region website:string is_active:bool

./bui generate price_group name:string code:string discount_percentage:float is_active:bool

./bui generate warehouse name:string code:string address_line1:string city:string country:belongsTo:Country is_active:bool

./bui generate product vinx_article_id:uint article_number:string sku:string name:translation slug:string price_unit:float vat_percentage:float is_active:bool

./bui generate product_price product:belongsTo:Product price_type:string price:float special_price:float special_price_from:datetime special_price_to:datetime customer_group_id:int qty:int is_active:bool priority:int

./bui generate stock_info article_id:int quantity_available:int quantity_current:int warehouse:belongsTo:Warehouse for_shop:bool

./bui generate stock_location_info stock_info:belongsTo:StockInfo location_code:string quantity:int

./bui generate configurable_product_option product:belongsTo:Product attribute:belongsTo:Attribute position:int

./bui generate configurable_product_link parent_product:belongsTo:Product child_product:belongsTo:Product

./bui generate bundle_option product:belongsTo:Product title:string type:string is_required:bool position:int

./bui generate bundle_option_selection bundle_option:belongsTo:BundleOption product:belongsTo:Product quantity:int position:int is_default:bool price_type:string price_value:float

./bui generate pricing_rule name:string description:text rule_type:string conditions:json actions:json priority:int is_active:bool valid_from:datetime valid_to:datetime

./bui generate product_section_code code:string name:translation description:text

./bui generate product_cycle name:string year:int harvest_date:datetime bottling_date:datetime

./bui generate customer name:string username:string first_name:string last_name:string email:string phone:string address_line1:string address_line2:string city:string postal_code:string country:belongsTo:Country region:belongsTo:Region notes:text status:string loyalty_points:int preferences:text password:string avatar:media:image reset_token:string reset_token_expiry:datetime last_login:datetime

./bui generate partner company_name:string trading_name:string tax_id:string vat_number:string business_type:string credit_limit:float payment_terms:int discount_percentage:float contact_person_first_name:string contact_person_last_name:string contact_person_email:string contact_person_phone:string primary_email:string primary_phone:string website:string notes:text status:string last_order_date:datetime total_order_value:float average_order_value:float price_group:belongsTo:PriceGroup account_manager:belongsTo:Employee

./bui generate partner_user partner:belongsTo:Partner user_role:belongsTo:UserRole email:string password:string first_name:string last_name:string username:string phone:string job_title:string department:string is_active:bool is_email_verified:bool last_login:datetime password_reset_token:string password_reset_expiry:datetime email_verification_token:string invitation_token:string invitation_expiry:datetime invited_by:belongsTo:Employee onboarding_completed:bool

./bui generate admin_user name:string username:string email:string avatar:media:image password:string last_login:datetime

./bui generate address addressable_type:string addressable_id:int type:string label:string first_name:string last_name:string company_name:string address_line1:string address_line2:string city:string state:string postal_code:string phone:string email:string is_default:bool is_active:bool country:belongsTo:Country region:belongsTo:Region

./bui generate credit partner:belongsTo:Partner credit_limit:float credit_used:float credit_available:float payment_terms:int credit_status:string last_credit_check:datetime credit_notes:text approved_by:belongsTo:Employee approved_at:datetime is_active:bool

./bui generate credit_request partner:belongsTo:Partner requested_amount:float current_limit:float requested_payment_terms:int business_justification:text financial_documents:json status:string requested_by:belongsTo:PartnerUser requested_at:datetime reviewed_by:belongsTo:Employee reviewed_at:datetime review_notes:text

./bui generate payment_term code:string name:translation days:int description:text is_active:bool

./bui generate order order_number:string order_date:datetime required_date:datetime shipped_date:datetime delivered_date:datetime status:string payment_status:string payment_method:string payment_terms:int subtotal:float discount_amount:float discount_percentage:float tax_amount:float tax_percentage:float shipping_cost:float total_amount:float notes:text internal_notes:text tracking_number:string shipping_method:string priority:string partner:belongsTo:Partner billing_address:belongsTo:Address shipping_address:belongsTo:Address created_by_user:belongsTo:Employee updated_by_user:belongsTo:Employee

./bui generate order_item order:belongsTo:Order product:belongsTo:Product quantity:int unit_price:float discount:float discount_type:string total_price:float notes:text

./bui generate rating designation:string minimal_rating:int maximal_rating:int unit:string

./bui generate product_rating product:belongsTo:Product rating:belongsTo:Rating score:float notes:text year:int source:string

./bui generate product_review product:belongsTo:Product customer:belongsTo:Customer rating:float review:text is_verified:bool is_approved:bool helpful_count:int

echo ""
echo "✓ All modules generated successfully!"
