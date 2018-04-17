package bot


/*

- interivews and hiring
- creating a workflow for doing an interview:

!workflow create carbon-interview
!workflow step add carbon-interview --expect candidate --expect time "Review resume and prep questions"    
!workflow step add carbon-interview "Do interview"
!workflow step add carbon-interview "Review interview stuff"

// this is ok, but cant' really execute reminder the way you'd like



*/



// Pipeline for making soup
/*
u: !workflow add soup
u: !workflow step add soup "Buy a can of your favorite soup"
u: !workflow step add soup "Put contents of soup can into bowl"
u: !workflow step add soup "Heat bowl in microwave for 1 minute"
u: !workflow step add soup "Eat soup"
# oops, forgot something
u: !workflow show soup
b: *Soup* Pipeline:
1. Buy a can of your favorite soup
2. Put contents of soup can into bowl
3. Heat bowl in microwave for 1 minute
4. Eat Soup

# oops, forgot something
u: !workflow add --index 4 soup "Wait for soup to cool off"
u: !workflow show soup
b: *Soup* Pipeline:
1. Buy a can of your favorite soup
2. Put contents of soup can into bowl
3. Heat bowl in microwave for 1 minute
4. Wait for soup to cool off
5. Eat Soup
*/

// Instance of someone using the soup workflow
/*
u: !thing start soup
u: !thing ls
b: Here are your things:
*Soup*

*/











// ok let's start from the users who are using the hring mechanism
// I'm a manager for team carbon, interviewing John Doe, using an existing workflow blueprint 
/*
u: !workflow ls
b: Here are the workflows I have: 
*Hire* 

*/



