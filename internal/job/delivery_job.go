package job

// описывает одну задачу доставки письма для конкретного почтальона
type DeliveryJob struct {
	ID        int
	PostmanID int
	MailText  string
	Address   string
	Priority  int
}
