import { Form, Input, InputNumber, Select, DatePicker, Button, Row, Col } from 'antd';
import dayjs from 'dayjs';

interface Props {
  initialValues?: any;
  onSubmit: (values: any) => void;
  onCancel: () => void;
}

export default function PersonEventForm({ initialValues, onSubmit, onCancel }: Props) {
  const [form] = Form.useForm();

  const handleFinish = (values: any) => {
    const data: any = {};
    const stringFields = [
      'name', 'attendance_group', 'id_number', 'gender', 'ethnicity',
      'native_place', 'address', 'political_status', 'marital_status', 'alias',
    ];
    const numFields = [
      'basic_salary', 'performance_salary', 'salary_days', 'position_allowance',
      'meal_subsidy', 'housing_subsidy', 'transport_subsidy', 'heat_subsidy',
      'insurance_compensation', 'housing_fund_compensation',
      'social_insurance_deduct', 'housing_fund_deduct',
    ];
    stringFields.forEach((f) => { if (values[f] !== undefined && values[f] !== null) data[f] = values[f]; });
    numFields.forEach((f) => { if (values[f] !== undefined && values[f] !== null) data[f] = Number(values[f]); });
    data.entry_date = values.entry_date ? dayjs(values.entry_date).format('YYYY-MM-DD') : undefined;
    data.birthday = values.birthday ? dayjs(values.birthday).format('YYYY-MM-DD') : undefined;
    data.phones = values.phones ? [String(values.phones)] : undefined;
    data.emails = values.emails ? [String(values.emails)] : undefined;
    data.bank_cards = values.bank_cards ? [String(values.bank_cards)] : undefined;

    onSubmit({
      effective_date: values.effective_date ? dayjs(values.effective_date).format('YYYY-MM-DD') : dayjs().format('YYYY-MM-DD'),
      event_type: values.event_type || 'info_change',
      data,
    });
  };

  const initial: any = initialValues || { salary_days: 21.75 };
  if (initial.effective_date) initial.effective_date = dayjs(initial.effective_date);
  if (initial.entry_date) initial.entry_date = dayjs(initial.entry_date);
  if (initial.birthday) initial.birthday = dayjs(initial.birthday);
  if (initial.phones && initial.phones.length > 0) initial.phones = initial.phones[0];
  if (initial.emails && initial.emails.length > 0) initial.emails = initial.emails[0];
  if (initial.bank_cards && initial.bank_cards.length > 0) initial.bank_cards = initial.bank_cards[0];

  return (
    <Form form={form} layout="vertical" initialValues={initial} onFinish={handleFinish} style={{ maxHeight: '65vh', overflowY: 'auto' }}>
      <Row gutter={16}>
        <Col span={12}><Form.Item name="effective_date" label="生效日期" rules={[{ required: true }]}><DatePicker style={{ width: '100%' }} /></Form.Item></Col>
        <Col span={12}>
          <Form.Item name="event_type" label="事件类型" rules={[{ required: true }]}>
            <Select options={[
              { label: '入职', value: 'onboard' }, { label: '调岗', value: 'position_change' },
              { label: '信息变更', value: 'info_change' }, { label: '离职', value: 'offboard' },
            ]} />
          </Form.Item>
        </Col>
        <Col span={8}><Form.Item name="name" label="姓名"><Input /></Form.Item></Col>
        <Col span={8}><Form.Item name="alias" label="别名"><Input /></Form.Item></Col>
        <Col span={8}><Form.Item name="attendance_group" label="考勤组"><Input /></Form.Item></Col>
        <Col span={8}><Form.Item name="entry_date" label="入职日期"><DatePicker style={{ width: '100%' }} /></Form.Item></Col>
        <Col span={8}><Form.Item name="basic_salary" label="基本工资"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
        <Col span={8}><Form.Item name="performance_salary" label="绩效工资"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
        <Col span={8}><Form.Item name="salary_days" label="计薪天数"><InputNumber style={{ width: '100%' }} min={1} /></Form.Item></Col>
        <Col span={8}><Form.Item name="position_allowance" label="职位津贴"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
        <Col span={8}><Form.Item name="meal_subsidy" label="餐补"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
        <Col span={8}><Form.Item name="housing_subsidy" label="房补"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
        <Col span={8}><Form.Item name="transport_subsidy" label="交通补贴"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
        <Col span={8}><Form.Item name="heat_subsidy" label="高温补贴"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
        <Col span={8}><Form.Item name="insurance_compensation" label="保险补偿"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
        <Col span={8}><Form.Item name="housing_fund_compensation" label="公积金补偿"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
        <Col span={8}><Form.Item name="social_insurance_deduct" label="社保代扣"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
        <Col span={8}><Form.Item name="housing_fund_deduct" label="公积金代扣"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
        <Col span={8}><Form.Item name="gender" label="性别"><Select options={[{ label: '男', value: '男' }, { label: '女', value: '女' }]} /></Form.Item></Col>
        <Col span={8}><Form.Item name="birthday" label="生日"><DatePicker style={{ width: '100%' }} /></Form.Item></Col>
        <Col span={8}><Form.Item name="id_number" label="身份证号"><Input /></Form.Item></Col>
        <Col span={8}><Form.Item name="phones" label="电话"><Input /></Form.Item></Col>
        <Col span={8}><Form.Item name="emails" label="邮箱"><Input /></Form.Item></Col>
        <Col span={8}><Form.Item name="ethnicity" label="民族"><Input /></Form.Item></Col>
        <Col span={8}><Form.Item name="native_place" label="籍贯"><Input /></Form.Item></Col>
        <Col span={16}><Form.Item name="address" label="住址"><Input /></Form.Item></Col>
        <Col span={8}><Form.Item name="bank_cards" label="银行卡号"><Input /></Form.Item></Col>
        <Col span={8}><Form.Item name="political_status" label="政治面貌"><Input /></Form.Item></Col>
        <Col span={8}><Form.Item name="marital_status" label="婚姻状态"><Select options={[{ label: '未婚', value: '未婚' }, { label: '已婚', value: '已婚' }]} /></Form.Item></Col>
      </Row>
      <div style={{ textAlign: 'right', marginTop: 16 }}>
        <Button style={{ marginRight: 8 }} onClick={onCancel}>取消</Button>
        <Button type="primary" htmlType="submit">保存</Button>
      </div>
    </Form>
  );
}
