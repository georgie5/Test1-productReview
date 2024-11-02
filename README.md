Owner: George Kotch
Test: 1 - productreviewapp

a. Create a product: 
    ```
    make addproduct name="" category="" image_url="" 
    ```

b. display a specific product
   ```
   curl -X GET http://localhost:4000/v1/products/1
   ``` 

   
c. update a specific product
    ```
    curl -X PATCH -H "Content-Type: application/json" -d '{"name":"update example"}' http://localhost:4000/v1/products/:productid
    ```

    
d. delete a specific product
     ```
     curl -X DELETE http://localhost:4000/v1/products/:productid
     ```

     
e. display all products
   ``` 
   curl -X GET http://localhost:4000/v1/products
   ```


f. Perform searching, filtering, sorting on products
    Query Parameters:
  ```
    name: Search by product name.
    category: Filter by category.
    sort: Sort by fields like name, category, etc. (use - prefix for descending).
    page: Specify page number for pagination.
    page_size: Specify the number of products per page.

    example: curl -X GET "http://localhost:4000/v1/products?name=example&category=Example&sort=-name&page=1&page_size=5"

  ```

g. create a review for a specific product
    ``` make addreview rating="" content="" productID="" ```

  
h. display a specific review for a specific product
    ``` curl -X GET http://localhost:4000/v1/products/:productid/reviews/:reviewid' ```

  
i. update a specific review for a specific product
    ``` curl -X PATCH -H "Content-Type: application/json" -d '{"rating":4,"content":"Updated review content"}' http://localhost:4000/v1/products/:productid/reviews/:reviewid ```

  
j. delete a specific review for a specific product
    ```curl -X DELETE http://localhost:4000/v1/products/:productid/reviews/:reviewid ``` 

    
k. display all reviews
    ``` curl -X GET http://localhost:4000/v1/reviews ```

    
l. display all reviews for a specific product
    ``` curl -X GET http://localhost:4000/v1/products/:productid/reviews ```

    
m. Perform searching, filtering, sorting on reviews

   Query Parameters: 
   ```
    rating: Filter reviews by rating.
    content: Search within review content.
    sort: Sort by fields like rating, helpful_count, etc. (use - prefix for descending).
    page: Specify page number for pagination.
    page_size: Specify the number of reviews per page.

    example: curl -X GET "http://localhost:4000/v1/reviews?rating=intexample&content=example&sort=helpful_count&page=1&page_size=5"

   ```

 additional: helpful_count
 
 ```curl -X POST http://localhost:4000/v1/products/:productid/reviews/:reviewid/helpful```
