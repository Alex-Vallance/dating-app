#My Dating App
###- Alex Vallance

###Intro
This is a simple dating app api which allows you register random users.
Once logged in, these users can get profiles of other users they might be interested in and 'swipe' indicating their interest.

###Setup
Open docker. Run **docker-compose build** and **docker-compose up**

Using postman, import the included in resources/DatingApp.postman_collection.json

Register users using the *create random user* request.
Remember to save the details for one or more of these in order to login.

Once logged in, you can copy the authentication as a bearer token to use the *get profiles for user* request.
 
Options are as follows

 *age_min*: ignored if less than 18 because that is the minimum age for the users

 *age_max*: ignored over 65 because that is the age max of the users

 *gender*: can specify 'Male' or 'Female' to restrict results

 *sort*:
- 'distance' will sort by users distance from the requesting user
- 'recommended' will sort users by their likability

Users are randomly assigned a location stored as latitude and longitude.

When a user requests profiles their distance is calculated from the requesting user.

*Likability* is determined by scoring how often users are liked and disliked by other users

A like is +1 and a dislike is -1

If a user is liked and then disliked by the same user (swiped then unmatched), their net score change will be 0.

Once a user has received their filtered profiles, they can 'swipe' on the user.

Remember the authentication.

To swipe a user they must provide the *profile_id* and 'YES' or 'NO'

If both users swipe 'YES', they will be matched and the response will have 'matched' and the profile id of the user they are matched with