package main

//type UserModel struct {
//	Id     int `xml:"id"`
//	FirstName   string `xml:"first_name"`
//	LastName   string `xml:"last_name"`
//	Age    int `xml:"age"`
//	About  string `xml:"about"`
//	Gender string `xml:"gender"`
//}
//
//type Users struct {
//	List    []UserModel `xml:"row"`
//}
//
//func (v *UserModel) Name() string{
//	return fmt.Sprintf("%s %s", v.FirstName, v.LastName)
//}
//
//func OrderByID(users []UserModel) []UserModel{
//	sort.Slice(users, func(i, j int) bool {
//		return users[i].Id < users[j].Id
//	})
//	return users
//}
//
//func OrderByAge(users []UserModel) []UserModel{
//	sort.Slice(users, func(i, j int) bool {
//		return users[i].Age < users[j].Age
//	})
//	return users
//}
//
//func OrderByName(users []UserModel) []UserModel{
//	sort.Slice(users, func(i, j int) bool {
//		return strings.Compare(users[i].Name(), users[j].Name()) < 0
//	})
//	return users
//}
//
//func Limit(users []UserModel, limit int, offset int) ([]UserModel, error) {
//	if offset >= len(users) || offset+limit >= len(users) {
//		return users, fmt.Errorf("limit or offset > users")
//	}
//	return users[offset : offset+limit], nil
//}
//
//func (v *Users) LoadXml(filename string){
//	xmlData, err := ioutil.ReadFile(filename)
//	if err != nil {
//		fmt.Printf("error: %v", err)
//		return
//	}
//
//	err = xml.Unmarshal(xmlData, &v)
//	if err != nil {
//		fmt.Printf("error: %v", err)
//		return
//	}
//}
//
//func (v *Users) Search(s string) []UserModel {
//	users := make([]UserModel, 0)
//	for _, u := range v.List {
//		if strings.Contains(u.About, s) || strings.Contains(u.Name(), s) {
//			users = append(users, u)
//		}
//	}
//	return users
//}
//
//
//func main() {
//	users := new(Users)
//	users.LoadXml("dataset.xml")
//
//	var err error
//	v := users.Search("")
//	v = OrderByName(v)
//	v, err = Limit(v, 5, 0)
//	if err != nil{
//		fmt.Println(err)
//	}
//
//
//
//	fmt.Printf("%+v\n", v)
//	fmt.Printf("len: %v\n", len(v))
//}


