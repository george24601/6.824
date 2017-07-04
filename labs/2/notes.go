for {
	//put select in a for loop, so that each msg in the channel we are looking for will be processed in this loop
	//note that time.After channel is refreshed in each for loop. That is why this implementation reprsents idle periods 
	select  {
	case msg := <- ch:
	{
		//process here
	}

	
	case <- time.After(time.Second):
	{
		//timeout logic here
	}


	}

}

//if a struct is used in RPC, make sure it is field name starts with UPPER CASE, otherwise it will not be serialized!


//make sure declare channel with buffer capacity, otherwise it will be blocking!
