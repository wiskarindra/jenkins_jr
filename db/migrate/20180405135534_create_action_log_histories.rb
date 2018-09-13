class CreateActionLogHistories < ActiveRecord::Migration[5.1]
  def up
    create_table :action_log_histories do |t|
      t.integer :record_id
      t.string :record_type
      t.text :changes
      t.bigint :actor_id

      t.timestamps null: false
    end

    add_index :action_log_histories, [:record_id, :record_type], name: "index_action_log_histories_on_record"
  end

  def down
    drop_table :action_log_histories
  end
end
