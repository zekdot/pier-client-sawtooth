package contract

//func (broker *Broker) GetOutterMeta(state *state.BrokerState)(map[string]uint64, error) {
//	meta, err := broker.getMap(state, outterMeta)
//	if err != nil {
//		return nil, err
//	}
//	return meta, nil
//}
//
//func (broker *Broker) GetInnerMeta(state *state.BrokerState) (map[string]uint64, error) {
//	meta, err := broker.getMap(state, innerMeta)
//	if err != nil {
//		return nil, err
//	}
//	return meta, nil
//}
//
//func (broker *Broker) GetCallbackMeta(state *state.BrokerState) (map[string]uint64, error) {
//	meta, err := broker.getMap(state, callbackMeta)
//	if err != nil {
//		return nil, err
//	}
//	return meta, nil
//}
//
//func (broker *Broker) GetOutMessage(state *state.BrokerState, destchainID string, sequenceNum string)(*Event, error) {
//	key := outMsgKey(destchainID, sequenceNum)
//	v, err := state.GetMetaData(key)
//	if err != nil {
//		return nil, err
//	}
//	res := &Event{}
//	err = json.Unmarshal([]byte(v), res)
//	if err != nil {
//		return nil, err
//	}
//	return res, nil
//}
//
//func (broker *Broker) getInMessage(state *state.BrokerState, sourceChainID string, sequenceNum string) (*Event, error) {
//	key := inMsgKey(sourceChainID, sequenceNum)
//	v, err := state.GetMetaData(key)
//	if err != nil {
//		return nil, err
//	}
//	res := &Event{}
//	err = json.Unmarshal([]byte(v), res)
//	if err != nil {
//		return nil, err
//	}
//	return res, nil
//}
//
//func (broker *Broker) markInCounter(state *state.BrokerState, from string) error {
//	inMeta, err := broker.getMap(state, innerMeta)
//	if err != nil {
//		return err
//	}
//
//	inMeta[from]++
//	return broker.putMap(state, innerMeta, inMeta)
//}
//
//func (broker *Broker) markCallbackCounter(state *state.BrokerState, from string, index uint64) error {
//	meta, err := broker.getMap(state, callbackMeta)
//	if err != nil {
//		return err
//	}
//
//	meta[from] = index
//
//	return broker.putMap(state, callbackMeta, meta)
//}