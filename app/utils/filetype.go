package controllers


func LoadTypeConf() map[string]string{
	conf := make(map[string]string)
	f, err := os.Open("filetypes.json")
	if err != nil {
		return conf
	}
	defer f.Close()
	jsonData, err := ioutil.ReadAll(f) 
	if err != nil {
		return conf
	}
	
    err = json.Unmarshal(jsonData, &conf)
    if err != nil {
        return conf
    }
	return conf
}

