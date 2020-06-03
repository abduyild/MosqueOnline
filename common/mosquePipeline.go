package common

type MosquePipeline struct {
	EditPrayers      bool // add/delete prayer, f.ex. add cuma prayer
	EditCapacity     bool // edit capacity for mosque, for women and men
	GetRegistrations bool // get users which want to attend choosen prayer, confirm visited
	GetDate          bool // get users registered for a choosen date
}
